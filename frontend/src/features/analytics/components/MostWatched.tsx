import React from 'react';
import { motion } from 'framer-motion';
import { Eye, Crown } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useMostWatched } from '../../../hooks/useAnalytics';
import type { MostWatchedMovie } from '../types/analytics.types';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
  return n.toString();
}

/** Width % of the bar relative to max views in the list */
function barWidth(views: number, maxViews: number): string {
  if (maxViews === 0) return '0%';
  return `${Math.max(4, Math.round((views / maxViews) * 100))}%`;
}

// ─── Skeleton ─────────────────────────────────────────────────────────────────

const SkeletonRow: React.FC<{ i: number }> = ({ i }) => (
  <div
    className="flex items-center gap-4 py-3 px-4 animate-pulse"
    style={{ animationDelay: `${i * 60}ms` }}
  >
    <div className="w-6 h-4 bg-flixor-gray rounded flex-shrink-0" />
    <div className="w-10 h-14 bg-flixor-gray rounded flex-shrink-0" />
    <div className="flex-1 space-y-2">
      <div className="h-3.5 bg-flixor-gray rounded w-3/4" />
      <div className="h-2.5 bg-flixor-gray/60 rounded w-1/2" />
      <div className="h-1.5 bg-flixor-gray/40 rounded w-full mt-1" />
    </div>
    <div className="w-12 h-4 bg-flixor-gray rounded flex-shrink-0" />
  </div>
);

// ─── Movie Row ────────────────────────────────────────────────────────────────

interface RowProps {
  movie: MostWatchedMovie;
  maxViews: number;
  index: number;
}

const MovieRow: React.FC<RowProps> = ({ movie, maxViews, index }) => (
  <motion.div
    initial={{ opacity: 0, y: 12 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.3, delay: index * 0.045, ease: 'easeOut' }}
  >
    <Link
      to={`/movie/${movie.id}`}
      className="flex items-center gap-4 py-3 px-4 rounded-xl hover:bg-white/4 border border-transparent hover:border-white/6 transition-all duration-200 group"
    >
      {/* Rank */}
      <span
        className="text-sm font-black w-5 text-center flex-shrink-0 tabular-nums"
        style={{ color: movie.rank <= 3 ? '#e50914' : '#555' }}
      >
        {movie.rank === 1 ? <Crown size={16} className="text-flixor-red mx-auto" /> : movie.rank}
      </span>

      {/* Thumbnail */}
      <div className="w-10 h-14 rounded-lg overflow-hidden bg-flixor-gray flex-shrink-0">
        {movie.thumbnail ? (
          <img
            src={movie.thumbnail}
            alt={movie.title}
            className="w-full h-full object-cover group-hover:brightness-110 transition-all duration-200"
            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
          />
        ) : (
          <div className="w-full h-full bg-flixor-gray" />
        )}
      </div>

      {/* Info + progress bar */}
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-white truncate group-hover:text-flixor-red transition-colors duration-200">
          {movie.title}
        </p>
        <p className="text-xs text-flixor-lightGray mt-0.5">
          {movie.year} · {movie.genre.slice(0, 2).join(', ')}
        </p>
        {/* Progress bar */}
        <div className="mt-2 h-1 bg-white/8 rounded-full overflow-hidden">
          <motion.div
            initial={{ width: 0 }}
            animate={{ width: barWidth(movie.total_views, maxViews) }}
            transition={{ duration: 0.6, delay: index * 0.045 + 0.2, ease: 'easeOut' }}
            className="h-full rounded-full bg-flixor-red"
          />
        </div>
      </div>

      {/* View count */}
      <div className="flex items-center gap-1.5 flex-shrink-0 text-flixor-lightGray">
        <Eye size={13} />
        <span className="text-sm tabular-nums font-medium">{fmt(movie.total_views)}</span>
      </div>
    </Link>
  </motion.div>
);

// ─── Main Component ───────────────────────────────────────────────────────────

const MostWatched: React.FC = () => {
  const { data, loading, error } = useMostWatched({ limit: 10 });
  const maxViews = data?.movies[0]?.total_views ?? 1;

  return (
    <div className="rounded-2xl border border-white/6 bg-[#181818] overflow-hidden h-full">
      {/* Header */}
      <div className="flex items-center gap-3 px-6 py-5 border-b border-white/6">
        <div className="w-8 h-8 rounded-lg bg-yellow-500/10 flex items-center justify-center text-yellow-400">
          <Crown size={18} />
        </div>
        <div>
          <h2 className="text-lg font-bold text-white">Most Watched</h2>
          <p className="text-xs text-flixor-lightGray">All-time view leaderboard</p>
        </div>
      </div>

      {/* Body */}
      <div className="px-2 py-2 space-y-0.5">
        {error ? (
          <p className="p-4 text-sm text-flixor-red">{error}</p>
        ) : loading ? (
          [...Array(8)].map((_, i) => <SkeletonRow key={i} i={i} />)
        ) : data?.movies.length === 0 ? (
          <p className="p-8 text-center text-flixor-lightGray text-sm">No data available.</p>
        ) : (
          data?.movies.map((movie, i) => (
            <MovieRow key={movie.id} movie={movie} maxViews={maxViews} index={i} />
          ))
        )}
      </div>
    </div>
  );
};

export default MostWatched;