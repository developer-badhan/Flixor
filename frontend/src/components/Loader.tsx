import React from 'react';

/**
 * The loader component for the application.
 * @returns A React component that represents the loader of the application.
 * @component
 * @returns {React.FC}
 */
const Loader: React.FC = () => {
  return (
    <div className="flex h-screen items-center justify-center bg-flixor-dark">
      <div className="flex flex-col items-center gap-4">
        <div className="h-16 w-16 animate-spin rounded-full border-4 border-flixor-gray border-t-flixor-red"></div>
        <p className="text-flixor-lightGray font-medium">Loading...</p>
      </div>
    </div>
  );
};

export default Loader;
