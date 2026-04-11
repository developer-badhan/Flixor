import api from '../../../services/api';

export const getMovies = async () => {
  const response = await api.get('/movies');
  return response.data; // Ensure this matches backend 'data' field structure
};

export const searchMovies = async (query: string) => {
  const response = await api.get(`/movies/search`, { params: { q: query } });
  return response.data;
};

export const getMovieById = async (id: string) => {
  const response = await api.get(`/movies/${id}`);
  return response.data;
};
