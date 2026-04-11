import React from 'react';
import { useFetch } from '../hooks/useFetch';
import MovieCard from '../components/MovieCard';

const MoviesPage: React.FC = () => {
  const { data: moviesResp, loading } = useFetch<any>('/movies');
  const movies = moviesResp?.data || [];

  return (
    <div className="pt-24 px-8 min-h-screen bg-flixor-dark">
      <h2 className="text-3xl font-bold mb-8 text-white">All Movies</h2>
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {loading ? (
           [...Array(12)].map((_, i) => <div key={i} className="aspect-[2/3] bg-flixor-gray animate-pulse rounded"></div>)
        ) : movies.length > 0 ? (
           movies.map((movie: any) => (
             <MovieCard key={movie.id} id={movie.id} title={movie.title} posterUrl={movie.thumbnail_url || ''} />
           ))
        ) : (
           <p className="text-flixor-lightGray">No movies available.</p>
        )}
      </div>
    </div>
  );
};

export default MoviesPage;
