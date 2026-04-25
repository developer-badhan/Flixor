import { useState, useCallback } from 'react';
import { interactionService, type ReactionType } from '../services/interactionService';

/** 
 * Manages like/dislike toggle state for a single movie.
 * Toggle behaviour: clicking the active reaction deactivates it (sends the same request again
 * which the backend upserts — a future "remove reaction" endpoint would be cleaner,
 * but for now we track optimistic local state without a second request).
*/

interface UseReactionReturn {
  reaction: ReactionType | null;   // current active reaction, null = no reaction
  loading: boolean;
  like: () => Promise<void>;
  dislike: () => Promise<void>;
}

export function useReaction(movieId: string): UseReactionReturn {
  const [reaction, setReaction] = useState<ReactionType | null>(null);
  const [loading, setLoading]   = useState(false);

  const like = useCallback(async () => {
    if (loading) return;
    setLoading(true);
    try {
      await interactionService.likeMovie(movieId);
      /** Toggle: if already liked → clear; otherwise set like
       */
      setReaction(prev => (prev === 'like' ? null : 'like'));
    } catch {
      /** Silent fail — optimistic state not applied on error*/
    } finally {
      setLoading(false);
    }
  }, [movieId, loading]);

  const dislike = useCallback(async () => {
    if (loading) return;
    setLoading(true);
    try {
      await interactionService.dislikeMovie(movieId);
      // Toggle: if already disliked → clear; otherwise set dislike
      setReaction(prev => (prev === 'dislike' ? null : 'dislike'));
    } catch {
      // Silent fail
    } finally {
      setLoading(false);
    }
  }, [movieId, loading]);

  return { reaction, loading, like, dislike };
}