import api from '../../../services/api';

export const getWatchlist = async () => {
  const response = await api.get('/interactions/watchlist');
  return response.data;
};

export const addToWatchlist = async (id: string) => {
  const response = await api.post(`/interactions/watchlist/${id}`);
  return response.data;
};

export const removeFromWatchlist = async (id: string) => {
  const response = await api.delete(`/interactions/watchlist/${id}`);
  return response.data;
};

export const getHistory = async () => {
  const response = await api.get('/interactions/history');
  return response.data;
};

export const sendOtp = async () => {
  const response = await api.post('/user/send-otp');
  return response.data;
};

export const verifyOtp = async (otp: string) => {
  const response = await api.post('/user/verify-otp', { otp });
  return response.data;
};
