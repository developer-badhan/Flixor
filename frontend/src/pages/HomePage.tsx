import React from 'react';
import { motion } from 'framer-motion';
import { Play, Info } from 'lucide-react';
import MovieCard from '../components/MovieCard';
import { useFetch } from '../hooks/useFetch';

const HomePage: React.FC = () => {
  // Correct extraction from nested API response
  const { data: moviesResp, loading, error } = useFetch<any>('/movies');

  // correct (based on your current useFetch)
  const movies = Array.isArray(moviesResp)
    ? moviesResp
    : moviesResp?.data || [];
  const heroMovie = movies[0] || null;
  const trendingMovies = movies.slice(0, 10);
  const recommendedMovies = movies.slice(10, 20);

  console.log("moviesResp:", moviesResp);

  return (
    <div className="min-h-screen bg-flixor-dark">
      
      {/* Hero Section */}
      <div className="relative h-[80vh] w-full">
        {heroMovie ? (
          <div className="absolute inset-0">
            <img 
              src={'https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?q=80&w=2070'} 
              alt={heroMovie.title}
              className="w-full h-full object-cover opacity-80"
            />
            <div className="absolute inset-0 bg-gradient-to-t from-flixor-dark via-flixor-dark/50 to-transparent"></div>
            <div className="absolute inset-0 bg-gradient-to-r from-flixor-dark via-flixor-dark/50 to-transparent"></div>
          </div>
        ) : (
          <div className="absolute inset-0 bg-flixor-gray animate-pulse"></div>
        )}

        <div className="relative h-full flex flex-col justify-center px-8 md:px-16 max-w-4xl pt-20">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            <h1 className="text-5xl md:text-7xl font-bold mb-4 drop-shadow-lg">
              {heroMovie?.title || 'Amazing Movie Title'}
            </h1>

            <p className="text-lg md:text-xl text-flixor-lightGray mb-8 max-w-2xl drop-shadow">
              {heroMovie?.description || 'A brief description of this featured movie.'}
            </p>

            <div className="flex items-center gap-4">
              <button className="flex items-center gap-2 bg-white text-black px-6 py-3 rounded hover:bg-opacity-80 font-bold text-lg">
                <Play size={24} fill="currentColor" /> Play
              </button>

              <button className="flex items-center gap-2 bg-gray-500/50 text-white px-6 py-3 rounded hover:bg-gray-500/70 font-bold text-lg">
                <Info size={24} /> More Info
              </button>
            </div>
          </motion.div>
        </div>
      </div>

      {/* Error State */}
      {error && (
        <div className="text-center text-red-500 py-6">
          Failed to load movies.
        </div>
      )}

      {/* Movie Sections */}
      <div className="px-8 md:px-16 pb-20 -mt-20 relative z-10">
        
        {/* Trending */}
        <section className="mb-12">
          <h2 className="text-2xl font-bold mb-6 text-white">Trending Now</h2>

          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
            {loading ? (
              [...Array(5)].map((_, i) => (
                <div key={i} className="aspect-[2/3] bg-flixor-gray animate-pulse rounded"></div>
              ))
            ) : trendingMovies.length > 0 ? (
                trendingMovies.map((movie: any) => (
                  <MovieCard
                    key={movie.id}
                    id={movie.id}
                    title={movie.title}
                    posterUrl={
                      movie.thumbnail_url ||
                      `https://archive.org/services/img/${movie.identifier}`
                    }
                  />  
                ))
            ) : (
              <p className="text-flixor-lightGray">No movies found.</p>
            )}
          </div>
        </section>

        {/* Recommended */}
        <section>
          <h2 className="text-2xl font-bold mb-6 text-white">Recommended For You</h2>

          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
            {loading ? (
              [...Array(5)].map((_, i) => (
                <div key={i} className="aspect-[2/3] bg-flixor-gray animate-pulse rounded"></div>
              ))
            ) : recommendedMovies.length > 0 ? (
                recommendedMovies.map((movie: any) => (
                  <MovieCard
                    key={movie.id}
                    id={movie.id}
                    title={movie.title}
                    posterUrl={
                      movie.thumbnail_url ||
                      `https://archive.org/services/img/${movie.identifier}`
                    }
                  />
                ))
            ) : (
              <p className="text-flixor-lightGray">
                Add metrics to see recommendations.
              </p>
            )}
          </div>
        </section>
      </div>
    </div>
  );
};

export default HomePage;

