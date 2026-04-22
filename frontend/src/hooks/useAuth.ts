/**
 *
 * 1. If user login or sign up it will store access token and refresh token in localStorage.
 * 2. If user logout or logoutAll it will remove access token and refresh token from localStorage.
 * 3. If user refresh token is expired it will automatically refresh token and store in localStorage.
 */
export { useAuthContext as useAuth } from '../context/AuthContext';