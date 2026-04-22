// ─── Trending ────────────────────────────────────────────────────────────────

export type TrendingWindow = '1d' | '7d' | '30d';

/**
 * - BUG FIX: genre was typed as `string[]` in both TrendingMovie and MostWatchedMovie.
 * - MongoDB documents can have genre = null when the field was never set on a movie.
 * - TypeScript trusted the type and gave no warning — runtime crashed with:
 *   "Cannot read properties of null (reading 'slice')"
 * - Fix: type genre as `string[] | null` everywhere, so components are
 * - forced to guard against null before calling .slice() or .join().
 */

export interface TrendingMovie {
  id: string;
  title: string;
  genre: string[] | null;  
  thumbnail: string;
  year: string;
  views_in_window: number;
  total_views: number;
  rank: number;
}

export interface TrendingResponse {
  window: TrendingWindow;
  window_label: string;
  total_results: number;
  movies: TrendingMovie[];
}

//  Most Watched 
export interface MostWatchedMovie {
  id: string;
  title: string;
  genre: string[] | null; 
  thumbnail: string;
  year: string;
  total_views: number;
  rank: number;
}

export interface MostWatchedResponse {
  total_results: number;
  movies: MostWatchedMovie[];
}

//  Top Genres 
export interface GenreStat {
  genre: string;
  movie_count: number;
  total_views: number;
  rank: number;
}

export interface TopGenresResponse {
  total_results: number;
  genres: GenreStat[];
}

//  Platform Stats 
export interface PlatformStats {
  total_movies: number;
  total_users: number;
  total_views: number;
  total_watchlist: number;
  total_likes: number;
}

// Query Params
export interface TrendingParams {
  window?: TrendingWindow;
  limit?: number;
}

export interface TopParams {
  limit?: number;
}