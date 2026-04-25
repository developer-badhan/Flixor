
import api from './api';

/**
 * HTTP layer for GET /movie/stream/:id
 * This endpoint:
 *   1. Requires JWT auth (handled by api.ts interceptor)
 *   2. Validates movie ID as MongoDB ObjectID
 *   3. Atomically increments view_count in DB
 *   4. Returns stream URL + metadata
 *
 * Response HAS success envelope:
 *   { success: true, data: { movie_id, title, stream_url, ... } }
 *   → api.ts interceptor detects "success" field → unwraps → response.data = inner object
 *
 * IMPORTANT field names from backend StreamResponse struct:
 *   stream_url    (not streamUrl)
 *   thumbnail_url (not thumbnailUrl, not thumbnail)
 *   genre         (not genres — StreamResponse uses genre:[]string)
 *   view_count    (not viewCount)
 */

// Types of stream movies
export interface StreamInfo {
  movie_id: string;
  title: string;
  stream_url: string;
  thumbnail_url: string;
  genre: string[] | null;   // can be null — same DB issue as analytics
  year: string;
  director: string;
  view_count: number;
}

// Service of stream movies

export const streamService = {
  /**
   * GET /movie/stream/:id
   * Triggers view count increment. Call ONLY when user actually plays the movie.
   * Returns unwrapped StreamInfo (api.ts interceptor removes the success envelope).
   *
   * Error codes from backend:
   *   400 → invalid movie ID format
   *   401 → no/invalid JWT
   *   404 → movie not found
   *   422 → movie exists but has no stream URL
   *   500 → unexpected server error
   */
  getStream: async (movieId: string): Promise<StreamInfo> => {
    const response = await api.get<StreamInfo>(`/movie/stream/${movieId}`);
    // api.ts interceptor: response.data.success exists → unwraps → response.data = StreamInfo
    return response.data;
  },
};