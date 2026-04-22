import axios from 'axios';

/** 
 *  BUGS FIXED:
 *    1. Refresh body key: { refreshToken } → { refresh_token }
 *       Backend binding tag is json:"refresh_token" — camelCase was silently ignored.
 *    2. New token extraction: res.data?.data?.accessToken → res.data?.data?.access_token
 *       Backend TokenResponse has json:"access_token" (snake_case).
 *    3. Everything else (queue logic, retry logic) was correct — kept as-is.
 */

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:5000/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Attach access token to every outgoing request
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Token refresh queue (handles concurrent 401s gracefully)
let isRefreshing = false;
let failedQueue: {
  resolve: (token: string) => void;
  reject: (err: unknown) => void;
}[] = [];

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token!);
    }
  });
  failedQueue = [];
};

// Response interceptor — unwrap + auto-refresh on 401
api.interceptors.response.use(
  (response) => {
    // Unwrap backend envelope: { success, message, data: {...} } → data
    if (response.data && response.data.success !== undefined) {
      return { ...response, data: response.data.data };
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status !== 401) {
      return Promise.reject(error);
    }

    if (originalRequest._retry) {
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({
          resolve: (token: string) => {
            originalRequest.headers['Authorization'] = `Bearer ${token}`;
            resolve(api(originalRequest));
          },
          reject: (err: unknown) => reject(err),
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      const refresh_token = localStorage.getItem('refreshToken');

      // Use raw axios (not `api`) to avoid triggering this interceptor again
      const res = await axios.post(
        `${import.meta.env.VITE_API_URL || 'http://localhost:5000/api/v1'}/auth/refresh`,
        {
          refresh_token,            // FIX 1: was { refreshToken } — backend needs snake_case
        }
      );

      // FIX 2: was res.data?.data?.accessToken — backend sends access_token (snake_case)
      const newAccessToken: string = res.data?.data?.access_token;
      const newRefreshToken: string = res.data?.data?.refresh_token;

      // Persist both tokens
      localStorage.setItem('accessToken', newAccessToken);
      localStorage.setItem('refreshToken', newRefreshToken);

      api.defaults.headers.common['Authorization'] = `Bearer ${newAccessToken}`;

      processQueue(null, newAccessToken);

      originalRequest.headers['Authorization'] = `Bearer ${newAccessToken}`;
      return api(originalRequest);

    } catch (err) {
      processQueue(err, null);

      // Full logout — wipe storage and redirect
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      window.location.href = '/login';

      return Promise.reject(err);
    } finally {
      isRefreshing = false;
    }
  }
);

export default api;