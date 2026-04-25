import { useState, useCallback } from 'react';
import { streamService, type StreamInfo } from '../services/streamService';

/** 
 *  Fetches stream info ONLY when the user clicks Play.
 *  We do NOT auto-fetch on mount because:
 *   1. GET /movie/stream/:id increments view_count in the DB atomically.
 *      Calling it on page load would inflate views without an actual watch.
 *   2. It requires auth — lazy fetch avoids 401 on public movie pages.
 *  
 *  Usage:
 *   const { streamInfo, loading, error, fetchStream } = useStream();
 *   // On Play click:
 *   const info = await fetchStream(movieId);
 *   if (info) setIsPlaying(true);
*/

interface UseStreamReturn {
  streamInfo: StreamInfo | null;
  loading: boolean;
  error: string | null;
  fetchStream: (movieId: string) => Promise<StreamInfo | null>;
}

export function useStream(): UseStreamReturn {
  const [streamInfo, setStreamInfo] = useState<StreamInfo | null>(null);
  const [loading, setLoading]       = useState(false);
  const [error, setError]           = useState<string | null>(null);

  /**
   * Fetch stream info and return it.
   * Returns null if the fetch fails (caller should NOT set isPlaying in that case).
   */
  const fetchStream = useCallback(async (movieId: string): Promise<StreamInfo | null> => {
    setLoading(true);
    setError(null);
    try {
      const info = await streamService.getStream(movieId);
      setStreamInfo(info);
      return info;
    } catch (err: any) {
      // Map backend error codes to user-facing messages
      const status = err?.response?.status;
      const backendMsg = err?.response?.data?.error;
      const msg =
        status === 404 ? 'Movie not found' :
        status === 422 ? 'Stream not available for this movie' :
        status === 401 ? 'Please sign in to watch' :
        backendMsg || 'Failed to load stream';
      setError(msg);
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  return { streamInfo, loading, error, fetchStream };
}