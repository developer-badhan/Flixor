import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { TrendingUp, Eye, Clock } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useTrending, type TrendingWindow } from '../../../hooks/useAnalytics';
import type { TrendingMovie } from '../types/analytics.types';

/** 
 *   BUG FIX (MovieRow, info section):
 *   BEFORE: movie.genre.slice(0, 2).join(', ')
 *   AFTER:  (movie.genre ?? []).slice(0, 2).join(', ')
 *   
 *   Identical root cause as MostWatched.tsx — MongoDB returns genre: null
 *   for movies synced without genre data. `?? []` falls back to empty array,
 *   so the separator dot is also hidden when there are no genres.
*/

//  Window Tab Config 
const WINDOWS: { value: TrendingWindow; label: string }[] = [
  { value: '1d', label: '24 Hours' },
  { value: '7d', label: '7 Days' },
  { value: '30d', label: '30 Days' },
];

//  Format helpers 
function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
  return n.toString();
}

//  Rank badge color 
function rankStyle(rank: number): { bg: string; text: string } {
  if (rank === 1) return { bg: '#e50914', text: '#fff' };
  if (rank === 2) return { bg: '#9ca3af', text: '#000' };
  if (rank === 3) return { bg: '#92400e', text: '#fbbf24' };
  return { bg: '#1f1f1f', text: '#9ca3af' };
}

//  Skeleton Row 
const SkeletonRow: React.FC<{ i: number }> = ({ i }) => (
  <div
    className="flex items-center gap-4 p-4 rounded-xl bg-flixor-gray/20 animate-pulse"
    style={{ animationDelay: `${i * 60}ms` }}
  >
    <div className="w-8 h-8 rounded-lg bg-flixor-gray flex-shrink-0" />
    <div className="w-14 h-20 rounded-lg bg-flixor-gray flex-shrink-0" />
    <div className="flex-1 space-y-2">
      <div className="h-4 bg-flixor-gray rounded w-2/3" />
      <div className="h-3 bg-flixor-gray/60 rounded w-1/3" />
    </div>
    <div className="text-right space-y-2">
      <div className="h-4 bg-flixor-gray rounded w-16" />
      <div className="h-3 bg-flixor-gray/60 rounded w-12" />
    </div>
  </div>
);

//  Movie Row 
const MovieRow: React.FC<{ movie: TrendingMovie; index: number }> = ({ movie, index }) => {
  const rs = rankStyle(movie.rank);
  const genres = movie.genre ?? []; // BUG FIX: guard null before any array operation

  return (
    <motion.div
      initial={{ opacity: 0, x: -16 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.3, delay: index * 0.05, ease: 'easeOut' }}
    >
      <Link
        to={`/movie/${movie.id}`}
        className="flex items-center gap-4 p-4 rounded-xl border border-transparent hover:border-white/8 hover:bg-white/4 transition-all duration-200 group"
      >
        {/* Rank badge */}
        <div
          className="w-8 h-8 rounded-lg flex items-center justify-center text-xs font-black flex-shrink-0"
          style={{ backgroundColor: rs.bg, color: rs.text }}
        >
          {movie.rank}
        </div>

        {/* Thumbnail */}
        <div className="w-14 h-20 rounded-lg overflow-hidden bg-flixor-gray flex-shrink-0">
          {movie.thumbnail ? (
            <img
              src={movie.thumbnail}
              alt={movie.title}
              className="w-full h-full object-cover brightness-90 group-hover:brightness-100 transition-all duration-300"
              onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center bg-flixor-gray">
              <span className="text-xs text-flixor-lightGray text-center px-1 leading-tight">
                {movie.title}
              </span>
            </div>
          )}
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <p className="font-semibold text-white truncate group-hover:text-flixor-red transition-colors duration-200">
            {movie.title}
          </p>
          <p className="text-xs text-flixor-lightGray mt-0.5">
            {/*
              BUG FIX:
              BEFORE: movie.genre.slice(0, 2).join(', ')  → crash when genre is null
              AFTER:  genres (= movie.genre ?? [])        → safe empty array fallback

              Dot separator is also conditional — avoids "2019 · " with nothing after it.
            */}
            {movie.year}
            {genres.length > 0 && <> · {genres.slice(0, 2).join(', ')}</>}
          </p>
        </div>

        {/* View metrics */}
        <div className="text-right flex-shrink-0">
          <div className="flex items-center gap-1.5 justify-end text-flixor-red">
            <TrendingUp size={13} />
            <span className="text-sm font-bold tabular-nums">{fmt(movie.views_in_window)}</span>
          </div>
          <div className="flex items-center gap-1.5 justify-end text-flixor-lightGray mt-1">
            <Eye size={12} />
            <span className="text-xs tabular-nums">{fmt(movie.total_views)}</span>
          </div>
        </div>
      </Link>
    </motion.div>
  );
};

//  Main Component 
const TrendingMovies: React.FC = () => {
  const [activeWindow, setActiveWindow] = useState<TrendingWindow>('7d');
  const { data, loading, error } = useTrending({ window: activeWindow, limit: 10 });

  return (
    <div className="rounded-2xl border border-white/6 bg-[#181818] overflow-hidden">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 px-6 py-5 border-b border-white/6">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-flixor-red/15 flex items-center justify-center text-flixor-red">
            <TrendingUp size={18} />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Trending Now</h2>
            {data && (
              <p className="text-xs text-flixor-lightGray">
                {data.window_label} · {data.total_results} movies
              </p>
            )}
          </div>
        </div>

        {/* Window Tabs */}
        <div className="flex items-center gap-1 bg-black/40 rounded-lg p-1">
          {WINDOWS.map((w) => (
            <button
              key={w.value}
              onClick={() => setActiveWindow(w.value)}
              className="relative px-3 py-1.5 rounded-md text-xs font-semibold transition-all duration-200"
              style={{ color: activeWindow === w.value ? '#fff' : '#b3b3b3' }}
            >
              {activeWindow === w.value && (
                <motion.div
                  layoutId="trending-window-pill"
                  className="absolute inset-0 rounded-md bg-flixor-red"
                  transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                />
              )}
              <span className="relative z-10 flex items-center gap-1.5">
                <Clock size={11} />
                {w.label}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* Body */}
      <div className="px-3 py-3 space-y-1">
        {error ? (
          <p className="p-4 text-sm text-flixor-red">{error}</p>
        ) : loading ? (
          [...Array(8)].map((_, i) => <SkeletonRow key={i} i={i} />)
        ) : data?.movies.length === 0 ? (
          <p className="p-8 text-center text-flixor-lightGray text-sm">
            No trending data for this window yet.
          </p>
        ) : (
          <AnimatePresence mode="wait">
            <motion.div key={activeWindow} className="space-y-1">
              {data?.movies.map((movie, i) => (
                <MovieRow key={movie.id} movie={movie} index={i} />
              ))}
            </motion.div>
          </AnimatePresence>
        )}
      </div>

      {/* Footer legend */}
      <div className="flex items-center gap-6 px-6 py-4 border-t border-white/6">
        <div className="flex items-center gap-1.5 text-xs text-flixor-lightGray">
          <TrendingUp size={12} className="text-flixor-red" />
          <span>Views in window</span>
        </div>
        <div className="flex items-center gap-1.5 text-xs text-flixor-lightGray">
          <Eye size={12} />
          <span>All-time views</span>
        </div>
      </div>
    </div>
  );
};

export default TrendingMovies;