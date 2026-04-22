import api from '../../../services/api';

/**
 * 
 * src/features/auth/services/authService.ts
 * 
 * BUGS FIXED:
 * 1. response.data.accessToken  → response.data.access_token   (backend: json:"access_token")
 * 2. response.data.refreshToken → response.data.refresh_token  (backend: json:"refresh_token")
 * 3. logout body { refreshToken } → { refresh_token }          (backend binding: json:"refresh_token")
 * 
 * NOTE: Token saving (localStorage + context) is now done in the LOGIN PAGE component,
 * not here, so AuthContext.setTokens() can be called to update shared React state.
 * This service only handles the HTTP calls and returns raw data.
 */

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

// LOGIN 
// Returns the token payload. Caller (LoginPage) must call setTokens() on context.
export const login = async (credentials: {
  email: string;
  password: string;
}): Promise<TokenResponse> => {
  const response = await api.post<TokenResponse>('/auth/login', credentials);

  // api.ts interceptor already unwraps response.data.data → response.data
  // So response.data IS the TokenResponse object from backend.
  return response.data;
};

// REGISTER 
export const register = async (userData: {
  username: string;
  email: string;
  password: string;
}): Promise<TokenResponse> => {
  const response = await api.post<TokenResponse>('/auth/register', userData);
  return response.data;
};

//  LOGOUT (single device) 
// FIX: key changed from refreshToken → refresh_token
export const logout = async (): Promise<void> => {
  const refresh_token = localStorage.getItem('refreshToken');

  await api.post('/auth/logout', {
    refresh_token,          // matches backend: json:"refresh_token"
  });
};

//  LOGOUT ALL DEVICES 
// FIX: key changed from refreshToken → refresh_token
export const logoutAll = async (): Promise<void> => {
  const refresh_token = localStorage.getItem('refreshToken');

  await api.post('/auth/logout-all', {
    refresh_token,          // matches backend: json:"refresh_token"
  });
};