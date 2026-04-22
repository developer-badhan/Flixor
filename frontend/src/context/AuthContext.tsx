/**
 * useAuth() was a plain hook — every component got its OWN isolated state copy.
 * ProtectedRoute's copy never saw the token saved by authService.
 * A React Context shares ONE state instance across the entire app.
 */

import React, { createContext, useContext, useState, useCallback } from 'react';

interface AuthContextType {
  isAuthenticated: boolean;
  accessToken: string | null;
  setTokens: (accessToken: string, refreshToken: string) => void;
  clearTokens: () => void;
}

// 1. Create the context with a safe default
const AuthContext = createContext<AuthContextType | null>(null);

// 2. Provider — wrap your entire app with this once in main.tsx
export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // Initialize from localStorage so page refreshes don't log the user out
  const [accessToken, setAccessToken] = useState<string | null>(
    localStorage.getItem('accessToken')
  );

  // Called after login or register — saves BOTH tokens
  const setTokens = useCallback((newAccessToken: string, newRefreshToken: string) => {
    setAccessToken(newAccessToken);
    localStorage.setItem('accessToken', newAccessToken);
    localStorage.setItem('refreshToken', newRefreshToken);
  }, []);

  // Called on logout — wipes everything
  const clearTokens = useCallback(() => {
    setAccessToken(null);
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  }, []);

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: !!accessToken,
        accessToken,
        setTokens,
        clearTokens,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

// 3. Safe consumer hook — throws if used outside provider (catches wiring mistakes early)
export function useAuthContext(): AuthContextType {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuthContext must be used inside <AuthProvider>');
  }
  return ctx;
}