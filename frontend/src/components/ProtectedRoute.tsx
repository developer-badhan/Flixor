import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

/**
 * Props for the ProtectedRoute component.
 * @param children - The children components to render if the user is authenticated.
 * @returns {React.ReactNode}
 */
interface ProtectedRouteProps {
  children?: React.ReactNode;
}

/**
 * The protected route component for the application.
 * @returns A React component that represents the protected route of the application.
 * @component
 * @returns {React.FC<ProtectedRouteProps>}
 */
const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children }) => {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return children ? <>{children}</> : <Outlet />;
};

export default ProtectedRoute;
