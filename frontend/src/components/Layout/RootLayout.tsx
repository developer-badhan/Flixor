import React from 'react';
import { Outlet } from 'react-router-dom';
import Navbar from '../Navbar';

/**
 * The root layout component for the application.
 * @returns A React component that represents the root layout of the application.
 * @component
 * @returns {React.FC}
 */
const RootLayout: React.FC = () => {
  return (
    <div className="min-h-screen flex flex-col bg-flixor-dark text-white">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="py-8 text-center text-flixor-lightGray text-sm">
        <p>&copy; {new Date().getFullYear()} Flixor. All rights reserved.</p>
      </footer>
    </div>
  );
};

export default RootLayout;
