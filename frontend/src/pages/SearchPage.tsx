import React from 'react';
import { useSearchParams } from 'react-router-dom';
import MovieCard from '../components/MovieCard';
import { useFetch } from '../hooks/useFetch';

const SearchPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('title') || '';

  // The backend returns results inside 'movies' property instead of 'data' based on default pagination struct?
  // Let's use any and inspect. We've seen moviesResp usually has data or movies
  const { data: searchResp, loading, error } = useFetch<any>(query ? `/movies/search?title=${encodeURIComponent(query)}` : null);

  const results = searchResp?.movies || searchResp?.data || [];

  return (
    <div className="min-h-screen bg-flixor-dark pt-24 px-8 md:px-16 pb-20">
      <h2 className="text-3xl font-bold mb-8 text-white">
        Search Results for <span className="text-flixor-red">"{query}"</span>
      </h2>

      {!query ? (
        <p className="text-flixor-lightGray">Please enter a search term.</p>
      ) : loading ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6">
          {[...Array(12)].map((_, i) => (
            <div key={i} className="aspect-[2/3] bg-flixor-gray animate-pulse rounded"></div>
          ))}
        </div>
      ) : error ? (
        <p className="text-flixor-red text-lg">Failed to load search results.</p>
      ) : results.length > 0 ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6">
          {results.map((movie: any) => (
            <MovieCard 
              key={movie.id} 
              id={movie.id} 
              title={movie.title} 
              posterUrl={movie.thumbnail_url || ''} 
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-20">
          <p className="text-flixor-lightGray text-xl">No movies found matching "{query}"</p>
          <p className="text-flixor-gray mt-2">Try searching for different keywords or genres.</p>
        </div>
      )}
    </div>
  );
};

export default SearchPage;
