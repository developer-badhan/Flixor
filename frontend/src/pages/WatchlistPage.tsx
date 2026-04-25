import React from 'react';
import { motion } from 'framer-motion';
import { BookmarkX, Loader2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Play } from 'lucide-react';
import { useWatchlist } from '../hooks/useWatchlist';
import type { EnrichedWatchlistItem } from '../hooks/useWatchlist';

/** 
 * Full watchlist UI:
 *   - Fetches watchlist from backend (movie_id + added_at)
 *   - Enriches each item with title + thumbnail from /movies/:id
 *   - Renders responsive grid of MovieCards with a remove button overlay
 *   - Shows skeleton while loading, empty state when empty, error state on failure
*/


// ─── Skeleton card ────────────────────────────────────────────────────────────

const SkeletonCard: React.FC = () => (
  <div className="aspect-[2/3] rounded-lg bg-flixor-gray animate-pulse" />
);

// ─── Watchlist movie card (extends MovieCard with remove button) ──────────────

interface WatchlistCardProps {
  item: EnrichedWatchlistItem;
  onRemove: (movieId: string) => void;
  removing: boolean;
}

const WatchlistCard: React.FC<WatchlistCardProps> = ({ item, onRemove, removing }) => {
  const [imgError, setImgError] = React.useState(false);
  const showFallback = !item.thumbnail_url || imgError;

  return (
    <motion.div
      layout
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: removing ? 0.4 : 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.9 }}
      transition={{ duration: 0.2 }}
      className="relative group"
    >
      {/* Card image */}
      <Link to={`/movie/${item.movie_id}`}>
        <div className="aspect-[2/3] rounded-lg overflow-hidden bg-flixor-gray cursor-pointer">
          {showFallback ? (
            <div className="w-full h-full flex items-center justify-center p-4 text-center bg-black/40">
              <span className="text-sm font-bold text-white leading-tight">{item.title}</span>
            </div>
          ) : (
            <img
              src={item.thumbnail_url}
              alt={item.title}
              className="w-full h-full object-cover brightness-90 group-hover:brightness-50 transition-all duration-300"
              onError={() => setImgError(true)}
            />
          )}

          {/* Hover overlay */}
          <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity duration-300 rounded-lg flex items-center justify-center">
            <div className="flex flex-col items-center gap-2">
              <div className="bg-flixor-red p-3 rounded-full">
                <Play size={20} fill="white" />
              </div>
              <p className="text-xs font-bold text-center px-2 text-white line-clamp-2">{item.title}</p>
            </div>
          </div>
        </div>
      </Link>

      {/* Remove button — top-right corner, visible on hover */}
      <button
        onClick={(e) => { e.preventDefault(); onRemove(item.movie_id); }}
        disabled={removing}
        title="Remove from watchlist"
        className="
          absolute top-2 right-2 z-10
          bg-black/70 hover:bg-flixor-red
          text-white rounded-full p-1.5
          opacity-0 group-hover:opacity-100
          transition-all duration-200
          disabled:cursor-not-allowed
        "
      >
        {removing
          ? <Loader2 size={14} className="animate-spin" />
          : <BookmarkX size={14} />
        }
      </button>

      {/* Title below card */}
      <p className="mt-2 text-xs text-flixor-lightGray truncate px-0.5">{item.title}</p>
    </motion.div>
  );
};

// ─── Page ─────────────────────────────────────────────────────────────────────

const WatchlistPage: React.FC = () => {
  const { movies, loading, error, removeFromWatchlist } = useWatchlist();
  const [removingIds, setRemovingIds] = React.useState<Set<string>>(new Set());

  const handleRemove = async (movieId: string) => {
    setRemovingIds(prev => new Set(prev).add(movieId));
    try {
      await removeFromWatchlist(movieId);
    } finally {
      setRemovingIds(prev => {
        const next = new Set(prev);
        next.delete(movieId);
        return next;
      });
    }
  };

  return (
    <div className="min-h-screen bg-flixor-dark pt-24 pb-20 px-6 md:px-10">
      <div className="max-w-7xl mx-auto">

        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl md:text-4xl font-bold text-white tracking-tight">My List</h1>
          {!loading && !error && (
            <p className="text-flixor-lightGray text-sm mt-1">
              {movies.length === 0
                ? 'No movies saved yet'
                : `${movies.length} movie${movies.length === 1 ? '' : 's'} saved`}
            </p>
          )}
        </div>

        {/* Error state */}
        {error && (
          <div className="flex flex-col items-center justify-center py-24 text-center">
            <p className="text-flixor-red text-lg mb-2">Failed to load your list</p>
            <p className="text-flixor-lightGray text-sm">{error}</p>
          </div>
        )}

        {/* Loading skeleton */}
        {loading && (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {Array.from({ length: 12 }).map((_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        )}

        {/* Empty state */}
        {!loading && !error && movies.length === 0 && (
          <div className="flex flex-col items-center justify-center py-32 text-center">
            <div className="w-16 h-16 rounded-full bg-white/5 flex items-center justify-center mb-4">
              <BookmarkX size={28} className="text-flixor-lightGray" />
            </div>
            <h2 className="text-xl font-semibold text-white mb-2">Your list is empty</h2>
            <p className="text-flixor-lightGray text-sm mb-6 max-w-xs">
              Add movies you want to watch later and they'll appear here.
            </p>
            <Link
              to="/movies"
              className="bg-flixor-red hover:bg-flixor-hover text-white text-sm font-semibold px-6 py-2.5 rounded-lg transition-colors"
            >
              Browse movies
            </Link>
          </div>
        )}

        {/* Movie grid */}
        {!loading && !error && movies.length > 0 && (
          <motion.div
            layout
            className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4"
          >
            {movies.map(item => (
              <WatchlistCard
                key={item.movie_id}
                item={item}
                onRemove={handleRemove}
                removing={removingIds.has(item.movie_id)}
              />
            ))}
          </motion.div>
        )}

      </div>
    </div>
  );
};

export default WatchlistPage;