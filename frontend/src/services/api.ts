import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:5000/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('accessToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
}, (error) => Promise.reject(error));


let isRefreshing = false;
let failedQueue: any[] = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

api.interceptors.response.use(
  (response) => {
    // unwrap backend response
    if (response.data && response.data.success !== undefined) {
      return { ...response, data: response.data.data };
    }
    return response;
  },
  async (error) => {
    const originalRequest = error.config;

    // ❌ If not 401 → reject
    if (error.response?.status !== 401) {
      return Promise.reject(error);
    }

    // ❌ Prevent infinite loop
    if (originalRequest._retry) {
      return Promise.reject(error);
    }

    // 🔒 If refresh already happening → queue request
    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        failedQueue.push({
          resolve: (token: string) => {
            originalRequest.headers['Authorization'] = `Bearer ${token}`;
            resolve(api(originalRequest));
          },
          reject: (err: any) => reject(err),
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      const refreshToken = localStorage.getItem('refreshToken');

      // 🚀 Call refresh API
      const res = await axios.post(
        `http://localhost:5000/api/v1/auth/refresh`,
        { refreshToken }
      );

      const newToken = res.data?.data?.accessToken;

      // ✅ Save new token
      localStorage.setItem('accessToken', newToken);

      // ✅ Update default header
      api.defaults.headers.common['Authorization'] = `Bearer ${newToken}`;

      processQueue(null, newToken);

      // 🔁 Retry original request
      originalRequest.headers['Authorization'] = `Bearer ${newToken}`;
      return api(originalRequest);

    } catch (err) {
      processQueue(err, null);

      // ❌ Logout user
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
