import React from 'react';

const WatchlistPage: React.FC = () => {
  return (
    <div className="pt-24 px-8 min-h-screen">
      <h2 className="text-3xl font-bold mb-8">My Watchlist</h2>
      {/* Grid placeholder */}
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {/* Movie cards will go here */}
      </div>
    </div>
  );
};

export default WatchlistPage;
