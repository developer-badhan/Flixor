// ─── Trending ────────────────────────────────────────────────────────────────

export type TrendingWindow = '1d' | '7d' | '30d';

export interface TrendingMovie {
  id: string;
  title: string;
  genre: string[];
  thumbnail: string;    // analytics model uses `thumbnail`, NOT `thumbnail_url`
  year: string;
  views_in_window: number;
  total_views: number;
  rank: number;
}

export interface TrendingResponse {
  window: TrendingWindow;
  window_label: string;   // e.g. "Last 7 days"
  total_results: number;
  movies: TrendingMovie[];
}

// ─── Most Watched ─────────────────────────────────────────────────────────────

export interface MostWatchedMovie {
  id: string;
  title: string;
  genre: string[];
  thumbnail: string;
  year: string;
  total_views: number;
  rank: number;
}

export interface MostWatchedResponse {
  total_results: number;
  movies: MostWatchedMovie[];
}

// ─── Top Genres ───────────────────────────────────────────────────────────────

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

// ─── Platform Stats ───────────────────────────────────────────────────────────

export interface PlatformStats {
  total_movies: number;
  total_users: number;
  total_views: number;
  total_watchlist: number;
  total_likes: number;
}

// ─── Query Params ─────────────────────────────────────────────────────────────

export interface TrendingParams {
  window?: TrendingWindow;
  limit?: number;
}

export interface TopParams {
  limit?: number;
}