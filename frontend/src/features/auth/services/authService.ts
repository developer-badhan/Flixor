import api from '../../../services/api';

// LOGIN
export const login = async (credentials: any) => {
  const response = await api.post('/auth/login', credentials);

  // Store tokens here (service layer is OK for this)
  localStorage.setItem('accessToken', response.data.accessToken);
  localStorage.setItem('refreshToken', response.data.refreshToken);

  return response.data;
};

// REGISTER
export const register = async (userData: any) => {
  const response = await api.post('/auth/register', userData);
  return response.data;
};

// LOGOUT (single device)
export const logout = async () => {
  const refreshToken = localStorage.getItem('refreshToken');

  const response = await api.post('/auth/logout', {
    refreshToken,
  });

  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');

  return response.data;
};

// LOGOUT ALL DEVICES
export const logoutAll = async () => {
  const refreshToken = localStorage.getItem('refreshToken');

  const response = await api.post('/auth/logout-all', {
    refreshToken,
  });

  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');

  return response.data;
};