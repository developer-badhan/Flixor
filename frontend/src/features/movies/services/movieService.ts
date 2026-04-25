import api from '../../../services/api';

/**
 * Get all movies from the database.
 * @returns A promise that resolves to the list of movies.
 * @async
 * @function getMovies
 * @returns {Promise<Movie[]>}
 */
export const getMovies = async () => {
  const response = await api.get('/movies');
  return response.data;
};

/**
 * Search movies by query.
 * @param query - The search query.
 * @returns A promise that resolves to the list of movies matching the query.
 * @async
 * @function searchMovies
 * @returns {Promise<Movie[]>}
 */
export const searchMovies = async (query: string) => {
  const response = await api.get(`/movies/search`, { params: { q: query } });
  return response.data;
};

/**
 * Get movie by ID.
 * @param id - The movie ID.
 * @returns A promise that resolves to the movie with the given ID.
 * @async
 * @function getMovieById
 * @returns {Promise<Movie>}
 */
export const getMovieById = async (id: string) => {
  const response = await api.get(`/movies/${id}`);
  return response.data;
};
