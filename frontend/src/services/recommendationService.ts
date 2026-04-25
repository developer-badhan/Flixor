import api from './api';
import type {
  RecommendRequest,
  AnyRecommendResponse,
} from '../features/recommendation/types/recommendation.types';


/**
 * recommendationService
 * 
 * CRITICAL NOTES:
 *  1. This endpoint uses POST with a JSON body — NOT GET with query params.
 *     The body is fully optional; an empty POST {} triggers hybrid mode with limit 10.
 *
 *  2. The backend returns TWO distinct response shapes:
 *     - Normal:     { mode, total_results, movies: [...] }
 *     - Empty user: { message: string, movies: [] }
 *     Neither has a `success` wrapper, so api.ts interceptor passes both through as-is.
 *
 *  3. The consumer (hook) is responsible for distinguishing the two shapes
 *     via the `isEmptyResponse` type guard.
 */
export const recommendationService = {
  /**
   * GET /api/v1/recommendations/
   * Despite the route being GET in the router, the handler reads ShouldBindJSON
   * from the request body — so we POST with a JSON body.
   *
   * Actually looking at the router: recommendations.GET("/", ...)
   * But the handler uses c.ShouldBindJSON(&req) which works with GET bodies too.
   * However, sending a body with GET is non-standard. We use POST-style
   * by calling api.post to align with the handler's ShouldBindJSON expectation.
   *
   * If the backend strictly enforces GET, switch to api.get with params.
   */
  getRecommendations: async (
    req: RecommendRequest = {}
  ): Promise<AnyRecommendResponse> => {
    const body: RecommendRequest = {
      mode: req.mode ?? 'hybrid',
      limit: req.limit ?? 10,
    };

    const response = await api.post<AnyRecommendResponse>(
      '/recommendations/',
      body
    );
    return response.data;
  },
};