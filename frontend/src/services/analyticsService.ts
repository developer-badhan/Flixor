import api from './api';
import type {
  TrendingResponse,
  MostWatchedResponse,
  TopGenresResponse,
  PlatformStats,
  TrendingParams,
  TopParams,
} from '../features/analytics/types/analytics.types';

/**
 * analyticsService
 *
 * NOTE: All four analytics endpoints return raw JSON with NO envelope wrapper
 * (unlike auth/user endpoints which return { success, message, data }).
 * The api.ts interceptor only unwraps when response.data.success !== undefined,
 * so analytics responses pass through as-is → response.data IS the payload.
 */

export const analyticsService = {
  /**
   * GET /api/v1/analytics/trending
   * @param params  window: '1d'|'7d'|'30d', limit: 1–100
   */
  getTrending: async (params: TrendingParams = {}): Promise<TrendingResponse> => {
    const { window = '7d', limit = 10 } = params;
    const response = await api.get<TrendingResponse>('/analytics/trending', {
      params: { window, limit },
    });
    return response.data;
  },

  /**
   * GET /api/v1/analytics/most-watched
   * @param params  limit: 1–100
   */
  getMostWatched: async (params: TopParams = {}): Promise<MostWatchedResponse> => {
    const { limit = 10 } = params;
    const response = await api.get<MostWatchedResponse>('/analytics/most-watched', {
      params: { limit },
    });
    return response.data;
  },

  /**
   * GET /api/v1/analytics/top-genres
   * @param params  limit: 1–100
   */
  getTopGenres: async (params: TopParams = {}): Promise<TopGenresResponse> => {
    const { limit = 10 } = params;
    const response = await api.get<TopGenresResponse>('/analytics/top-genres', {
      params: { limit },
    });
    return response.data;
  },

  /**
   * GET /api/v1/analytics/stats
   * Returns flat PlatformStats object — no query params needed.
   */
  getPlatformStats: async (): Promise<PlatformStats> => {
    const response = await api.get<PlatformStats>('/analytics/stats');
    return response.data;
  },
};