import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Play, Plus, Check, ThumbsUp, ThumbsDown, Loader2 } from 'lucide-react';
import { useFetch } from '../hooks/useFetch';
import { useStream } from '../hooks/useStream';
import { useReaction } from '../hooks/useReaction';
import { interactionService } from '../services/interactionService';

/**
 * All four buttons are now wired to real API calls:
 * 
 * Play button:
 *   BEFORE: directly used movie.stream_url (from /movies/:id — no view count increment)
 *   AFTER:  calls GET /movie/stream/:id on click → increments view_count in DB → gets
 *           the real stream_url from StreamResponse → plays video
 *   WHY:    The stream endpoint is the correct one. The /movies/:id endpoint may expose
 *           stream_url in the Movie model, but using /movie/stream/:id ensures
 *           view tracking, auth enforcement, and correct DTO shape.
 * 
 * + (watchlist) button:
 *   BEFORE: no API call
 *   AFTER:  POST /interactions/watchlist/:id → toggles in/out of watchlist
 *   
 * ThumbsUp (like) button:
 *   BEFORE: no API call
 *   AFTER:  POST /interactions/like/:id → body { reaction: "like" }
 * 
 * ThumbsDown (dislike) button:
 *   BEFORE: Heart icon with no API call
 *   AFTER:  ThumbsDown icon → POST /interactions/dislike/:id → body { reaction: "dislike" }
 */

const MovieDetailsPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { data: movie, loading, error } = useFetch<any>(`/movies/${id}`);

  // ── Stream state ──────────────────────────────────────────────────────────
  const { streamInfo, loading: streamLoading, error: streamError, fetchStream } = useStream();
  const [isPlaying, setIsPlaying] = useState(false);

  // ── Watchlist state ───────────────────────────────────────────────────────
  const [inWatchlist, setInWatchlist] = useState(false);
  const [watchlistLoading, setWatchlistLoading] = useState(false);

  // ── Reaction state ────────────────────────────────────────────────────────
  const { reaction, loading: reactionLoading, like, dislike } = useReaction(id ?? '');

  // ── Handlers ─────────────────────────────────────────────────────────────

  const handlePlay = async () => {
    if (!id) return;
    if (streamInfo) {
      // Already fetched — play immediately (view count already incremented)
      setIsPlaying(true);
      return;
    }
    const info = await fetchStream(id);
    if (info) {
      setIsPlaying(true);
    }
  };

  const handleWatchlist = async () => {
    if (!id || watchlistLoading) return;
    setWatchlistLoading(true);
    try {
      if (inWatchlist) {
        await interactionService.removeFromWatchlist(id);
        setInWatchlist(false);
      } else {
        await interactionService.addToWatchlist(id);
        setInWatchlist(true);
      }
    } catch {
      // Silent fail — button state stays unchanged on error
    } finally {
      setWatchlistLoading(false);
    }
  };

  // ── Loading / error states ────────────────────────────────────────────────

  if (loading) {
    return (
      <div className="h-screen bg-flixor-dark flex items-center justify-center">
        <Loader2 size={32} className="animate-spin text-flixor-lightGray" />
      </div>
    );
  }

  if (error || !movie) {
    return (
      <div className="h-screen bg-flixor-dark flex items-center justify-center text-red-500">
        Error loading movie
      </div>
    );
  }

  // ── Video player view ─────────────────────────────────────────────────────

  if (isPlaying) {
    // Use stream URL from StreamResponse (fetched via /movie/stream/:id)
    // Fall back to movie.stream_url if somehow already set (safety net only)
    const url = streamInfo?.stream_url || movie.stream_url || '';

    return (
      <div className="w-full h-screen bg-black flex items-center justify-center pt-20 relative">
        {url ? (
          <video
            className="w-full h-full object-contain"
            controls
            autoPlay
            src={url}
            aria-label="Video Player"
          >
            Your browser does not support the video tag.
          </video>
        ) : (
          <div className="text-white text-center">
            <p className="text-lg mb-2">Stream not available</p>
            <p className="text-flixor-lightGray text-sm">This movie has no video source.</p>
          </div>
        )}
        <button
          className="absolute top-24 right-8 bg-white/20 hover:bg-white/40 text-white px-4 py-2 rounded-lg z-50 transition-colors text-sm font-medium"
          onClick={() => setIsPlaying(false)}
        >
          ✕ Close
        </button>
      </div>
    );
  }

  // ── Detail view ────────────────────────────────────────────────────────────

  return (
    <div className="min-h-screen bg-flixor-dark">
      <div className="relative pb-20">

        {/* Backdrop */}
        <div className="relative h-[60vh] w-full">
          <img
            src="https://images.unsplash.com/photo-1626814026160-2237a95fc5a0?q=80&w=2070"
            alt={movie.title}
            className="w-full h-full object-cover opacity-30"
          />
          <div className="absolute inset-0 bg-gradient-to-t from-flixor-dark to-transparent" />
        </div>

        {/* Content */}
        <div className="max-w-6xl mx-auto px-8 -mt-32 relative z-10">
          <motion.div
            className="flex flex-col md:flex-row gap-8"
            initial={{ opacity: 0, y: 50 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            {/* Poster */}
            <img
              src={movie.thumbnail_url || 'https://via.placeholder.com/300x450'}
              alt={movie.title}
              className="w-64 rounded-lg shadow-2xl border border-flixor-gray hidden md:block"
            />

            {/* Info */}
            <div className="flex-1">
              <h1 className="text-4xl md:text-5xl font-bold mb-4">{movie.title}</h1>

              <div className="flex items-center gap-4 text-flixor-lightGray mb-6 text-sm font-medium">
                <span>{movie.year || '2023'}</span>
                <span className="border border-flixor-gray px-1 rounded">{movie.age_rating || '16+'}</span>
                <span>{movie.duration ? `${movie.duration} min` : '120 min'}</span>
                <span className="text-flixor-red font-bold">HD</span>
              </div>

              <p className="text-lg mb-8 leading-relaxed max-w-3xl">{movie.description}</p>

              {/* Stream error message */}
              {streamError && (
                <p className="text-red-400 text-sm mb-4">{streamError}</p>
              )}

              {/* Action buttons */}
              <div className="flex items-center gap-4">

                {/* Play */}
                <button
                  className="flex items-center gap-2 bg-white text-black px-8 py-3 rounded font-bold text-lg hover:bg-opacity-80 transition-all disabled:opacity-60"
                  onClick={handlePlay}
                  disabled={streamLoading}
                >
                  {streamLoading
                    ? <Loader2 size={20} className="animate-spin" />
                    : <Play fill="currentColor" size={20} />
                  }
                  {streamLoading ? 'Loading...' : 'Play'}
                </button>

                {/* + Watchlist toggle */}
                {/* BEFORE: static button with no handler */}
                {/* AFTER: calls addToWatchlist / removeFromWatchlist, shows check when added */}
                <button
                  onClick={handleWatchlist}
                  disabled={watchlistLoading}
                  title={inWatchlist ? 'Remove from list' : 'Add to list'}
                  className="p-3 border-2 border-gray-500 rounded-full hover:border-white transition-colors bg-[#2a2a2a]/60 backdrop-blur disabled:opacity-50"
                >
                  {watchlistLoading
                    ? <Loader2 size={20} className="animate-spin" />
                    : inWatchlist
                      ? <Check size={20} className="text-green-400" />
                      : <Plus size={20} />
                  }
                </button>

                {/* Like */}
                {/* BEFORE: no handler */}
                {/* AFTER: POST /interactions/like/:id with body { reaction: "like" } */}
                <button
                  onClick={like}
                  disabled={reactionLoading}
                  title="Like"
                  className={`
                    p-3 border-2 rounded-full transition-colors bg-[#2a2a2a]/60 backdrop-blur disabled:opacity-50
                    ${reaction === 'like'
                      ? 'border-flixor-red text-flixor-red'
                      : 'border-gray-500 hover:border-white'
                    }
                  `}
                >
                  <ThumbsUp size={20} />
                </button>

                {/* Dislike */}
                {/* BEFORE: Heart icon, no handler */}
                {/* AFTER: ThumbsDown → POST /interactions/dislike/:id with { reaction: "dislike" } */}
                <button
                  onClick={dislike}
                  disabled={reactionLoading}
                  title="Dislike"
                  className={`
                    p-3 border-2 rounded-full transition-colors bg-[#2a2a2a]/60 backdrop-blur disabled:opacity-50
                    ${reaction === 'dislike'
                      ? 'border-blue-400 text-blue-400'
                      : 'border-gray-500 hover:border-white'
                    }
                  `}
                >
                  <ThumbsDown size={20} />
                </button>

              </div>

              <div className="mt-8 space-y-2 text-sm">
                <p>
                  <span className="text-flixor-lightGray">Genres: </span>
                  {/* genres (plural) from /movies/:id Movie model, not genre from StreamResponse */}
                  {(movie.genres ?? []).join(', ') || 'Unknown'}
                </p>
                <p>
                  <span className="text-flixor-lightGray">Director: </span>
                  {movie.director || 'Unknown'}
                </p>
              </div>

            </div>
          </motion.div>
        </div>

      </div>
    </div>
  );
};

export default MovieDetailsPage;