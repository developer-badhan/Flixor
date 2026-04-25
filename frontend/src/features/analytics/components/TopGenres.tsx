import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { BarChart2, Film, Eye } from 'lucide-react';
import { useTopGenres } from '../../../hooks/useAnalytics';
import type { GenreStat } from '../types/analytics.types';

/**
 * Formats a number with appropriate suffixes (K, M, etc.).
 * @param n - The number to format.
 * @returns The formatted number string.
 * @function fmt
 * @returns {string}
 */
function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
  return n.toString();
}

function barPct(value: number, max: number): number {
  if (max === 0) return 0;
  return Math.max(3, Math.round((value / max) * 100));
}

// A small palette of distinct hues for the genre bars
const GENRE_COLORS = [
  '#e50914', // red
  '#3b82f6', // blue
  '#10b981', // green
  '#f59e0b', // amber
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#84cc16', // lime
  '#f97316', // orange
  '#a855f7', // violet
];

type Metric = 'total_views' | 'movie_count';

// ─── Skeleton ─────────────────────────────────────────────────────────────────

const SkeletonRow: React.FC<{ i: number }> = ({ i }) => (
  <div
    className="flex items-center gap-3 animate-pulse"
    style={{ animationDelay: `${i * 50}ms` }}
  >
    <div className="w-4 h-3 bg-flixor-gray rounded flex-shrink-0" />
    <div className="w-20 h-3 bg-flixor-gray rounded flex-shrink-0" />
    <div className="flex-1 h-6 bg-flixor-gray/40 rounded-full" />
    <div className="w-12 h-3 bg-flixor-gray rounded flex-shrink-0" />
  </div>
);

/**
 * Interface for the genre row component.
 * @interface GenreRowProps
 * @property {GenreStat} genre - The genre to display.
 * @property {number} max - The maximum number of views.
 * @property {Metric} metric - The metric to display.
 * @property {number} index - The index of the genre row.
 * @returns {React.FC<GenreRowProps>}
 */
interface GenreRowProps {
  genre: GenreStat;
  max: number;
  metric: Metric;
  index: number;
}

const GenreRow: React.FC<GenreRowProps> = ({ genre, max, metric, index }) => {
  const color = GENRE_COLORS[index % GENRE_COLORS.length];
  const value = metric === 'total_views' ? genre.total_views : genre.movie_count;
  const pct = barPct(value, max);

  return (
    <motion.div
      initial={{ opacity: 0, x: 12 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ duration: 0.3, delay: index * 0.04, ease: 'easeOut' }}
      className="flex items-center gap-3 group"
    >
      {/* Rank */}
      <span className="text-xs font-bold w-4 text-center tabular-nums" style={{ color: index < 3 ? color : '#555' }}>
        {genre.rank}
      </span>

      {/* Genre name */}
      <span className="text-xs font-semibold text-flixor-lightGray w-20 truncate flex-shrink-0 group-hover:text-white transition-colors">
        {genre.genre}
      </span>

      {/* Bar */}
      <div className="flex-1 h-6 bg-white/5 rounded-full overflow-hidden">
        <motion.div
          initial={{ width: 0 }}
          animate={{ width: `${pct}%` }}
          transition={{ duration: 0.55, delay: index * 0.04 + 0.15, ease: 'easeOut' }}
          className="h-full rounded-full flex items-center justify-end pr-2 relative"
          style={{ backgroundColor: `${color}30`, borderRight: `2px solid ${color}` }}
        >
          {/* Glow tip */}
          <div
            className="absolute right-0 top-0 bottom-0 w-4 rounded-r-full"
            style={{ background: `linear-gradient(to right, transparent, ${color}60)` }}
          />
        </motion.div>
      </div>

      {/* Value */}
      <span
        className="text-xs font-bold tabular-nums w-12 text-right flex-shrink-0"
        style={{ color }}
      >
        {fmt(value)}
      </span>
    </motion.div>
  );
};

// ─── Main Component ───────────────────────────────────────────────────────────

const TopGenres: React.FC = () => {
  const [metric, setMetric] = useState<Metric>('total_views');
  const { data, loading, error } = useTopGenres({ limit: 10 });

  const maxValue = metric === 'total_views'
    ? (data?.genres[0]?.total_views ?? 1)
    : (data?.genres.reduce((a, g) => Math.max(a, g.movie_count), 1) ?? 1);

  return (
    <div className="rounded-2xl border border-white/6 bg-[#181818] overflow-hidden h-full">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-6 py-5 border-b border-white/6">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-400">
            <BarChart2 size={18} />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white">Top Genres</h2>
            <p className="text-xs text-flixor-lightGray">Ranked by viewership</p>
          </div>
        </div>

        {/* Metric toggle */}
        <div className="flex items-center gap-1 bg-black/40 rounded-lg p-1">
          <button
            onClick={() => setMetric('total_views')}
            className="relative flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-semibold transition-all duration-200"
            style={{ color: metric === 'total_views' ? '#fff' : '#b3b3b3' }}
          >
            {metric === 'total_views' && (
              <motion.div
                layoutId="genre-metric-pill"
                className="absolute inset-0 rounded-md bg-blue-600"
                transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              />
            )}
            <span className="relative z-10 flex items-center gap-1.5">
              <Eye size={11} /> Views
            </span>
          </button>
          <button
            onClick={() => setMetric('movie_count')}
            className="relative flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-semibold transition-all duration-200"
            style={{ color: metric === 'movie_count' ? '#fff' : '#b3b3b3' }}
          >
            {metric === 'movie_count' && (
              <motion.div
                layoutId="genre-metric-pill"
                className="absolute inset-0 rounded-md bg-blue-600"
                transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              />
            )}
            <span className="relative z-10 flex items-center gap-1.5">
              <Film size={11} /> Movies
            </span>
          </button>
        </div>
      </div>

      {/* Body */}
      <div className="px-6 py-4 space-y-3">
        {error ? (
          <p className="text-sm text-flixor-red">{error}</p>
        ) : loading ? (
          [...Array(8)].map((_, i) => <SkeletonRow key={i} i={i} />)
        ) : data?.genres.length === 0 ? (
          <p className="py-8 text-center text-flixor-lightGray text-sm">No genre data available.</p>
        ) : (
          <motion.div key={metric} className="space-y-3">
            {data?.genres.map((genre, i) => (
              <GenreRow key={genre.genre} genre={genre} max={maxValue} metric={metric} index={i} />
            ))}
          </motion.div>
        )}
      </div>

      {/* Footer */}
      <div className="px-6 py-4 border-t border-white/6 flex items-center justify-between">
        <p className="text-xs text-flixor-lightGray">
          {data ? `${data.total_results} genres tracked` : ''}
        </p>
        <div className="flex items-center gap-1.5 text-xs text-flixor-lightGray">
          <span>Sorted by {metric === 'total_views' ? 'total views' : 'movie count'}</span>
        </div>
      </div>
    </div>
  );
};

export default TopGenres;