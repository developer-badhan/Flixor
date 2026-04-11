import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:5000/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
}, (error) => Promise.reject(error));

api.interceptors.response.use((response) => {
  // Our backend wraps responses in { success: true, message: string, data: any }
  if (response.data && response.data.success !== undefined) {
    // Overwrite axios's response.data with our actual payload
    return { ...response, data: response.data.data };
  }
  return response;
}, (error) => Promise.reject(error));

export default api;
