import React from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Play, Eye, Quote } from 'lucide-react';
import type { RecommendedMovie } from '../types/recommendation.types';
import SourceBadge from './SourceBadge';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function fmtViews(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M views';
  if (n >= 1_000)     return (n / 1_000).toFixed(1) + 'K views';
  return `${n} views`;
}

// ─── Genre pill ───────────────────────────────────────────────────────────────

const GenrePill: React.FC<{ label: string }> = ({ label }) => (
  <span
    className="px-2 py-0.5 rounded-md text-xs font-semibold"
    style={{
      background: 'rgba(255,255,255,0.07)',
      color: '#9ca3af',
      border: '1px solid rgba(255,255,255,0.08)',
    }}
  >
    {label}
  </span>
);

// ─── Skeleton card ────────────────────────────────────────────────────────────

export const RecommendationCardSkeleton: React.FC<{ index: number }> = ({ index }) => (
  <div
    className="rounded-2xl overflow-hidden bg-[#1a1a1a] border border-white/5 animate-pulse"
    style={{ animationDelay: `${index * 70}ms` }}
  >
    {/* Thumbnail placeholder */}
    <div className="aspect-[2/3] bg-[#242424]" />
    {/* Body */}
    <div className="p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div className="h-3 w-20 bg-[#2a2a2a] rounded" />
        <div className="h-5 w-16 bg-[#2a2a2a] rounded-md" />
      </div>
      <div className="h-4 w-3/4 bg-[#2a2a2a] rounded" />
      <div className="flex gap-1.5">
        <div className="h-4 w-12 bg-[#2a2a2a] rounded-md" />
        <div className="h-4 w-14 bg-[#2a2a2a] rounded-md" />
      </div>
      <div className="space-y-1.5 pt-1 border-t border-white/5">
        <div className="h-3 w-full bg-[#2a2a2a] rounded" />
        <div className="h-3 w-2/3 bg-[#2a2a2a] rounded" />
      </div>
    </div>
  </div>
);

// ─── Main card ────────────────────────────────────────────────────────────────

interface RecommendationCardProps {
  movie: RecommendedMovie;
  index: number;
}

const RecommendationCard: React.FC<RecommendationCardProps> = ({ movie, index }) => {
  const isTopPick = movie.source === 'rule+ai';

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, delay: index * 0.055, ease: 'easeOut' }}
      className="group relative rounded-2xl overflow-hidden border flex flex-col"
      style={{
        background: '#1a1a1a',
        borderColor: isTopPick
          ? 'rgba(229,9,20,0.25)'
          : 'rgba(255,255,255,0.06)',
      }}
    >
      {/* Top-pick shimmer border */}
      {isTopPick && (
        <div
          className="absolute inset-0 rounded-2xl pointer-events-none z-10"
          style={{
            background:
              'linear-gradient(135deg, rgba(229,9,20,0.08) 0%, transparent 50%)',
          }}
        />
      )}

      {/* ── Thumbnail ── */}
      <Link to={`/movie/${movie.id}`} className="relative aspect-[2/3] overflow-hidden bg-[#242424] flex-shrink-0">
        {movie.thumbnail ? (
          <img
            src={movie.thumbnail}
            alt={movie.title}
            className="w-full h-full object-cover brightness-90 group-hover:brightness-75 transition-all duration-400"
            onError={(e) => {
              (e.target as HTMLImageElement).style.display = 'none';
            }}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-[#1f1f1f]">
            <span className="text-xs text-[#444] text-center px-4 font-medium">
              {movie.title}
            </span>
          </div>
        )}

        {/* Play overlay */}
        <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300">
          <motion.div
            initial={{ scale: 0.8 }}
            whileHover={{ scale: 1 }}
            className="w-12 h-12 rounded-full flex items-center justify-center"
            style={{ background: 'rgba(229,9,20,0.9)', backdropFilter: 'blur(4px)' }}
          >
            <Play size={20} fill="white" className="text-white ml-0.5" />
          </motion.div>
        </div>

        {/* Source badge — pinned top-right */}
        <div className="absolute top-2.5 right-2.5 z-20">
          <SourceBadge source={movie.source} />
        </div>

        {/* Year — pinned bottom-left */}
        <div
          className="absolute bottom-2.5 left-2.5 text-xs font-bold px-2 py-0.5 rounded"
          style={{
            background: 'rgba(0,0,0,0.75)',
            backdropFilter: 'blur(4px)',
            color: '#9ca3af',
          }}
        >
          {movie.year}
        </div>
      </Link>

      {/* ── Card body ── */}
      <div className="flex flex-col flex-1 p-4 gap-3">

        {/* Title + views row */}
        <div>
          <Link
            to={`/movie/${movie.id}`}
            className="text-sm font-bold text-white leading-snug line-clamp-2 hover:text-[#e50914] transition-colors duration-200 block"
          >
            {movie.title}
          </Link>
          <div
            className="flex items-center gap-1 mt-1"
            style={{ color: '#4b5563' }}
          >
            <Eye size={11} />
            <span className="text-xs">{fmtViews(movie.view_count)}</span>
          </div>
        </div>

        {/* Genre pills */}
        {movie.genre.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {movie.genre.slice(0, 3).map((g) => (
              <GenrePill key={g} label={g} />
            ))}
          </div>
        )}

        {/* Match reason — the core UX value */}
        <div
          className="flex gap-2 pt-3 border-t mt-auto"
          style={{ borderColor: 'rgba(255,255,255,0.05)' }}
        >
          <Quote
            size={12}
            className="flex-shrink-0 mt-0.5"
            style={{ color: isTopPick ? '#e50914' : '#4b5563' }}
          />
          <p
            className="text-xs leading-relaxed italic"
            style={{ color: '#6b7280' }}
          >
            {movie.match_reason}
          </p>
        </div>
      </div>
    </motion.div>
  );
};

export default RecommendationCard;