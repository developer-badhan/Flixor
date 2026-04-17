import { useState, useEffect, useCallback } from 'react';
import { analyticsService } from '../services/analyticsService';
import type {
  TrendingResponse,
  MostWatchedResponse,
  TopGenresResponse,
  PlatformStats,
  TrendingParams,
  TopParams,
  TrendingWindow,
} from '../features/analytics/types/analytics.types';

// ─── Generic fetch state shape ────────────────────────────────────────────────

interface FetchState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

// ─── 1. useTrending ───────────────────────────────────────────────────────────
/**
 * Hook for GET /analytics/trending
 * Re-fetches automatically when `window` param changes (tab switcher).
 */
export function useTrending(params: TrendingParams = {}): FetchState<TrendingResponse> {
  const { window = '7d', limit = 10 } = params;
  const [data, setData] = useState<TrendingResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await analyticsService.getTrending({ window, limit });
      setData(result);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load trending movies');
    } finally {
      setLoading(false);
    }
  }, [window, limit]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

// ─── 2. useMostWatched ────────────────────────────────────────────────────────

export function useMostWatched(params: TopParams = {}): FetchState<MostWatchedResponse> {
  const { limit = 10 } = params;
  const [data, setData] = useState<MostWatchedResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await analyticsService.getMostWatched({ limit });
      setData(result);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load most-watched movies');
    } finally {
      setLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

// ─── 3. useTopGenres ──────────────────────────────────────────────────────────

export function useTopGenres(params: TopParams = {}): FetchState<TopGenresResponse> {
  const { limit = 10 } = params;
  const [data, setData] = useState<TopGenresResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await analyticsService.getTopGenres({ limit });
      setData(result);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load top genres');
    } finally {
      setLoading(false);
    }
  }, [limit]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

// ─── 4. usePlatformStats ──────────────────────────────────────────────────────

export function usePlatformStats(): FetchState<PlatformStats> {
  const [data, setData] = useState<PlatformStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetch = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await analyticsService.getPlatformStats();
      setData(result);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load platform stats');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetch();
  }, [fetch]);

  return { data, loading, error, refetch: fetch };
}

// ─── Re-export window type for components ─────────────────────────────────────
export type { TrendingWindow };