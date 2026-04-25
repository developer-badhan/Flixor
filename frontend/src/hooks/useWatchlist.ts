import { useState, useEffect, useCallback } from 'react';
import { interactionService } from '../services/interactionService';
import api from '../services/api';

/** 
 * Manages watchlist state for the WatchlistPage.
 * 
 *  Problem: Backend watchlist only returns { movie_id, added_at }.
 *  WatchlistPage needs title + thumbnail to render MovieCard.
 *  Solution: fetch watchlist → enrich each item with /movies/:id in parallel.
 * 
 *  Also exposes: addToWatchlist, removeFromWatchlist, isInWatchlist
 *  for use in MovieDetailsPage (add/remove button).
*/

export interface EnrichedWatchlistItem {
  movie_id: string;
  added_at: string;
  title: string;
  thumbnail_url: string;
}

interface UseWatchlistReturn {
  movies: EnrichedWatchlistItem[];
  loading: boolean;
  error: string | null;
  isInWatchlist: (movieId: string) => boolean;
  addToWatchlist: (movieId: string) => Promise<void>;
  removeFromWatchlist: (movieId: string) => Promise<void>;
  refetch: () => void;
}

// Hook
export function useWatchlist(): UseWatchlistReturn {
  const [movies, setMovies] = useState<EnrichedWatchlistItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAndEnrich = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // Step 1: Get watchlist — returns { watchlist: [{movie_id, added_at}], count }
      // No success envelope — response.data IS the payload (see interactionService)
      const { watchlist } = await interactionService.getWatchlist();

      if (!watchlist || watchlist.length === 0) {
        setMovies([]);
        return;
      }

      /**
       * Step 2: Enrich each item with movie details in parallel
       * /movies/:id is a public endpoint — no auth needed for this lookup
       */
      const enriched = await Promise.all(
        watchlist.map(async (item) => {
          try {
            const res = await api.get(`/movies/${item.movie_id}`);
            /**
             *  /movies/:id may or may not have success envelope depending on implementation
             *  api.ts will unwrap if it has one — either way res.data has the movie fields
             */
            const m = res.data ?? {};
            return {
              movie_id: item.movie_id,
              added_at: item.added_at,
              title: m.title ?? 'Unknown Title',
              thumbnail_url: m.thumbnail_url ?? '',
            };
          } catch {
            // Movie might have been deleted — keep the item with a placeholder
            return {
              movie_id: item.movie_id,
              added_at: item.added_at,
              title: 'Unknown Title',
              thumbnail_url: '',
            };
          }
        })
      );

      setMovies(enriched);
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Failed to load watchlist');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAndEnrich();
  }, [fetchAndEnrich]);

  // Watchlist mutation helpers
  const addToWatchlist = useCallback(async (movieId: string) => {
    await interactionService.addToWatchlist(movieId);
    // Don't re-enrich the full list — let MovieDetailsPage manage its own button state
  }, []);

  const removeFromWatchlist = useCallback(async (movieId: string) => {
    await interactionService.removeFromWatchlist(movieId);
    // Optimistic update — remove from local state immediately
    setMovies(prev => prev.filter(m => m.movie_id !== movieId));
  }, []);

  const isInWatchlist = useCallback(
    (movieId: string) => movies.some(m => m.movie_id === movieId),
    [movies]
  );

  return {
    movies,
    loading,
    error,
    isInWatchlist,
    addToWatchlist,
    removeFromWatchlist,
    refetch: fetchAndEnrich,
  };
}