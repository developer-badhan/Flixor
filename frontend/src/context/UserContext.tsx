import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { userService, type UserProfile } from '../services/userService';
import { useAuthContext } from './AuthContext';

/**
 * UserContextType is a TypeScript interface that defines the shape of the user context.
 * It contains the user's profile information, loading state, and methods to refresh and set the user data.
 * 
 * @property {UserProfile | null} user - The user's profile information
 * @property {boolean} loading - Indicates if the user data is being loaded
 * @property {() => Promise<void>} refreshUser - Method to refresh the user data
 * @property {React.Dispatch<React.SetStateAction<UserProfile | null>>} setUser - Method to set the user data
 */

interface UserContextType {
  user: UserProfile | null;
  loading: boolean;
  refreshUser: () => Promise<void>;
  setUser: React.Dispatch<React.SetStateAction<UserProfile | null>>;
}

/**
 * The user context for the application.
 * @returns A React component that represents the user context of the application.
 * @component
 * @returns {React.FC}
 */
const UserContext = createContext<UserContextType | null>(null);

/**
 * UserProvider is a React component that provides user context to its children.
 * It fetches and manages the user's profile information.
 * 
 * @param {React.ReactNode} children - The child components that need access to the user context
 * @returns {JSX.Element} - The UserProvider component
 */
export const UserProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAuthContext();
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const refreshUser = useCallback(async () => {
    if (!isAuthenticated) {
      setUser(null);
      setLoading(false);
      return;
    }
    
    setLoading(true);
    try {
      const userData = await userService.getMe();
      setUser(userData);
    } catch (error) {
      console.error('Failed to fetch user profile:', error);
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, [isAuthenticated]);

  useEffect(() => {
    refreshUser();
  }, [refreshUser]);

  return (
    <UserContext.Provider
      value={{
        user,
        loading,
        refreshUser,
        setUser,
      }}
    >
      {children}
    </UserContext.Provider>
  );
};

/**
 * Hook that provides access to the user context.
 * @returns {UserContextType} - The user context
 * @throws {Error} - Throws an error if the hook is used outside of the UserProvider
 */
export function useUserContext(): UserContextType {
  const ctx = useContext(UserContext);
  if (!ctx) {
    throw new Error('useUserContext must be used inside <UserProvider>');
  }
  return ctx;
}
