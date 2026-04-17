import { useState, useEffect } from 'react';

export function useAuth() {
  const [token, setToken] = useState<string | null>(
    localStorage.getItem('accessToken')
  );

  const isAuthenticated = !!token;

  const login = (accessToken: string) => {
    setToken(accessToken);
    localStorage.setItem('accessToken', accessToken);
  };

  const logout = () => {
    setToken(null);
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  };

  const logoutAll = () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    setToken(null);
  };

  return { isAuthenticated, token, login, logout, logoutAll };
}