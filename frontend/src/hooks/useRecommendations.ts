import { useState, useCallback } from 'react';
import { recommendationService } from '../services/recommendationService';
import {
  isEmptyResponse,
} from '../features/recommendation/types/recommendation.types';
import type {
  RecommendMode,
  RecommendedMovie,
} from '../features/recommendation/types/recommendation.types';

// ─── Hook State Shape ─────────────────────────────────────────────────────────

export interface RecommendationState {
  movies: RecommendedMovie[];
  mode: RecommendMode;
  totalResults: number;
  isNewUser: boolean;         // true when backend returns the empty-state shape
  emptyMessage: string;       // populated when isNewUser === true
  loading: boolean;
  error: string | null;
  fetch: (mode?: RecommendMode, limit?: number) => void;
}

// ─── Hook ─────────────────────────────────────────────────────────────────────

/**
 * useRecommendations
 *
 * Drives the recommendation page. Call `fetch(mode)` to trigger or re-trigger.
 * Does NOT auto-fetch on mount — the page component calls fetch() in a useEffect
 * so the user sees the mode selector before data loads.
 *
 * Key behaviours:
 *  - Detects the new-user empty-state shape and sets isNewUser = true
 *  - Mode parameter drives the POST body; changing mode re-fetches
 *  - Errors from both engines in hybrid mode surface as a single error string
 */
export function useRecommendations(): RecommendationState {
  const [movies, setMovies]             = useState<RecommendedMovie[]>([]);
  const [mode, setMode]                 = useState<RecommendMode>('hybrid');
  const [totalResults, setTotalResults] = useState(0);
  const [isNewUser, setIsNewUser]       = useState(false);
  const [emptyMessage, setEmptyMessage] = useState('');
  const [loading, setLoading]           = useState(false);
  const [error, setError]               = useState<string | null>(null);

  const fetch = useCallback(async (
    requestedMode: RecommendMode = 'hybrid',
    limit = 10,
  ) => {
    setLoading(true);
    setError(null);
    setIsNewUser(false);
    setEmptyMessage('');
    setMode(requestedMode);

    try {
      const data = await recommendationService.getRecommendations({
        mode: requestedMode,
        limit,
      });

      if (isEmptyResponse(data)) {
        // New user — backend returned { message, movies: [] }
        setIsNewUser(true);
        setEmptyMessage(data.message);
        setMovies([]);
        setTotalResults(0);
      } else {
        // Normal response — { mode, total_results, movies }
        setMovies(data.movies);
        setTotalResults(data.total_results);
        setIsNewUser(false);
      }
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.response?.data?.detail ||
        'Failed to load recommendations';
      setError(msg);
      setMovies([]);
      setTotalResults(0);
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    movies,
    mode,
    totalResults,
    isNewUser,
    emptyMessage,
    loading,
    error,
    fetch,
  };
}