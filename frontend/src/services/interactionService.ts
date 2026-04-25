
import api from './api';

/** 
 *  HTTP layer for all interaction endpoints.
 *  Key rule: Watchlist and History responses have NO success envelope.
 *   → api.ts interceptor only unwraps when response.data.success !== undefined
 *   → these pass through raw → read response.data.watchlist / response.data.history directly
 *  
 *  React (like/dislike) responses have NO envelope either:
 *   → { message: "reaction saved", reaction: "like" }
 *  
 *  Route param is :movieId (after router.go bug fix).
*/

// Types of interactions
export interface WatchlistItem {
  movie_id: string;
  added_at: string;
}

export interface WatchlistResponse {
  watchlist: WatchlistItem[];
  count: number;
}

export interface WatchEvent {
  movie_id: string;
  watched_at: string;
}

export interface HistoryResponse {
  history: WatchEvent[];
  count: number;
}

export type ReactionType = 'like' | 'dislike';

// Service of interactions

export const interactionService = {

  // Watchlist
  /**
   * GET /interactions/watchlist
   * Response shape: { watchlist: [{movie_id, added_at}], count: N }
   * NO success envelope — read response.data directly.
   */
  getWatchlist: async (): Promise<WatchlistResponse> => {
    const response = await api.get<WatchlistResponse>('/interactions/watchlist');
    // api.ts interceptor does NOT unwrap (no "success" field) → response.data IS the payload
    return response.data;
  },

  /**
   * POST /interactions/watchlist/:movieId
   * No body. Returns { message: "movie added to watchlist" }
   */
  addToWatchlist: async (movieId: string): Promise<void> => {
    await api.post(`/interactions/watchlist/${movieId}`);
  },

  /**
   * DELETE /interactions/watchlist/:movieId
   * No body. Returns { message: "movie removed from watchlist" }
   */
  removeFromWatchlist: async (movieId: string): Promise<void> => {
    await api.delete(`/interactions/watchlist/${movieId}`);
  },

  // Reactions
  /**
   * POST /interactions/like/:movieId
   * Body: { reaction: "like" }  ← backend reads from JSON body, not path
   * Returns { message: "reaction saved", reaction: "like" }
   */
  likeMovie: async (movieId: string): Promise<void> => {
    await api.post(`/interactions/like/${movieId}`, { reaction: 'like' });
  },

  /**
   * POST /interactions/dislike/:movieId
   * Body: { reaction: "dislike" }
   * Returns { message: "reaction saved", reaction: "dislike" }
   */
  dislikeMovie: async (movieId: string): Promise<void> => {
    await api.post(`/interactions/dislike/${movieId}`, { reaction: 'dislike' });
  },

  // Watch History
  /**
   * GET /interactions/history
   * Response shape: { history: [{movie_id, watched_at}], count: N }
   * NO success envelope — read response.data directly.
   */
  getHistory: async (): Promise<HistoryResponse> => {
    const response = await api.get<HistoryResponse>('/interactions/history');
    return response.data;
  },
};