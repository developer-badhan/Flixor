import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Sparkles, RefreshCw, Layers, Cpu } from 'lucide-react';
import { useRecommendations } from '../hooks/useRecommendations';
import {
  ModeSelector,
  RecommendationCard,
  RecommendationCardSkeleton,
  EmptyState,
} from '../features/recommendation';
import type { RecommendMode } from '../features/recommendation/types/recommendation.types';

// ─── Mode label map ───────────────────────────────────────────────────────────

const MODE_LABELS: Record<RecommendMode, { label: string; icon: React.ReactNode; color: string }> = {
  hybrid:  { label: 'Hybrid Mode',      icon: <Layers size={13} />,   color: '#e50914' },
  rule:    { label: 'Genre Matching',   icon: <Cpu size={13} />,      color: '#3b82f6' },
  ai:      { label: 'Gemini AI',        icon: <Sparkles size={13} />, color: '#a855f7' },
};

// ─── Results header ───────────────────────────────────────────────────────────

interface ResultsHeaderProps {
  totalResults: number;
  mode: RecommendMode;
  onRefresh: () => void;
  loading: boolean;
}

const ResultsHeader: React.FC<ResultsHeaderProps> = ({
  totalResults,
  mode,
  onRefresh,
  loading,
}) => {
  const modeInfo = MODE_LABELS[mode];
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-3">
        <span className="text-sm font-bold text-white">
          {totalResults} {totalResults === 1 ? 'pick' : 'picks'} for you
        </span>
        <div
          className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold"
          style={{
            background: `${modeInfo.color}12`,
            color: modeInfo.color,
            border: `1px solid ${modeInfo.color}25`,
          }}
        >
          {modeInfo.icon}
          {modeInfo.label}
        </div>
      </div>
      <button
        onClick={onRefresh}
        disabled={loading}
        className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all duration-200 hover:bg-white/5 disabled:opacity-40"
        style={{ color: '#6b7280', border: '1px solid rgba(255,255,255,0.06)' }}
      >
        <RefreshCw size={12} className={loading ? 'animate-spin' : ''} />
        Refresh
      </button>
    </div>
  );
};

// ─── Error banner ─────────────────────────────────────────────────────────────

const ErrorBanner: React.FC<{ message: string; onRetry: () => void }> = ({
  message,
  onRetry,
}) => (
  <motion.div
    initial={{ opacity: 0, y: -8 }}
    animate={{ opacity: 1, y: 0 }}
    className="flex items-center justify-between gap-4 px-5 py-4 rounded-xl"
    style={{
      background: 'rgba(229,9,20,0.08)',
      border: '1px solid rgba(229,9,20,0.2)',
    }}
  >
    <p className="text-sm text-red-400">{message}</p>
    <button
      onClick={onRetry}
      className="text-xs font-bold text-red-400 hover:text-red-300 transition-colors flex-shrink-0"
    >
      Try again
    </button>
  </motion.div>
);

// ─── Page ─────────────────────────────────────────────────────────────────────

const RecommendationPage: React.FC = () => {
  const {
    movies,
    mode,
    totalResults,
    isNewUser,
    emptyMessage,
    loading,
    error,
    fetch,
  } = useRecommendations();

  const [activeMode, setActiveMode] = useState<RecommendMode>('hybrid');

  // Initial fetch on mount
  useEffect(() => {
    fetch('hybrid', 20);
  }, [fetch]);

  const handleModeChange = (newMode: RecommendMode) => {
    setActiveMode(newMode);
    fetch(newMode, 20);
  };

  const handleRefresh = () => {
    fetch(activeMode, 20);
  };

  return (
    <div className="min-h-screen bg-flixor-dark pt-24 pb-20">

      {/* Ambient background glow */}
      <div
        className="fixed top-0 left-1/2 -translate-x-1/2 w-[600px] h-[300px] pointer-events-none z-0 opacity-10"
        style={{
          background:
            'radial-gradient(ellipse, rgba(229,9,20,0.6) 0%, rgba(168,85,247,0.3) 40%, transparent 70%)',
          filter: 'blur(60px)',
        }}
      />

      <div className="relative z-10 max-w-7xl mx-auto px-6 md:px-10 space-y-8">

        {/* ── Page header ── */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, ease: 'easeOut' }}
        >
          <div className="flex items-start gap-4">
            <div
              className="w-11 h-11 rounded-xl flex items-center justify-center flex-shrink-0 mt-0.5"
              style={{
                background: 'rgba(229,9,20,0.12)',
                border: '1px solid rgba(229,9,20,0.2)',
              }}
            >
              <Sparkles size={22} style={{ color: '#e50914' }} />
            </div>
            <div>
              <h1 className="text-3xl md:text-4xl font-bold text-white tracking-tight leading-none">
                For You
              </h1>
              <p className="text-sm mt-1.5" style={{ color: '#6b7280' }}>
                Personalised picks powered by genre analysis &amp; Gemini AI
              </p>
            </div>
          </div>

          {/* Separator */}
          <div
            className="mt-6 h-px"
            style={{
              background:
                'linear-gradient(to right, rgba(229,9,20,0.35), rgba(168,85,247,0.15), transparent)',
            }}
          />
        </motion.div>

        {/* ── Mode selector ── */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, delay: 0.1, ease: 'easeOut' }}
        >
          <p
            className="text-xs font-semibold uppercase tracking-widest mb-3"
            style={{ color: '#4b5563' }}
          >
            Recommendation Engine
          </p>
          <ModeSelector
            activeMode={activeMode}
            loading={loading}
            onChange={handleModeChange}
          />
        </motion.div>

        {/* ── Error banner ── */}
        <AnimatePresence>
          {error && (
            <ErrorBanner message={error} onRetry={handleRefresh} />
          )}
        </AnimatePresence>

        {/* ── Results section ── */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, delay: 0.2, ease: 'easeOut' }}
          className="space-y-5"
        >
          {/* Results header — only show when we have results */}
          {!loading && !isNewUser && movies.length > 0 && (
            <ResultsHeader
              totalResults={totalResults}
              mode={mode}
              onRefresh={handleRefresh}
              loading={loading}
            />
          )}

          {/* ── Empty state — new user ── */}
          {!loading && !error && isNewUser && (
            <EmptyState message={emptyMessage} />
          )}

          {/* ── Movie grid ── */}
          <AnimatePresence mode="wait">
            {loading ? (
              // Skeleton grid
              <motion.div
                key="skeletons"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4"
              >
                {[...Array(10)].map((_, i) => (
                  <RecommendationCardSkeleton key={i} index={i} />
                ))}
              </motion.div>
            ) : !isNewUser && movies.length > 0 ? (
              // Results grid
              <motion.div
                key={`results-${activeMode}`}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.25 }}
                className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4"
              >
                {movies.map((movie, i) => (
                  <RecommendationCard key={movie.id} movie={movie} index={i} />
                ))}
              </motion.div>
            ) : null}
          </AnimatePresence>
        </motion.div>

        {/* ── How it works — footer note ── */}
        {!loading && !isNewUser && movies.length > 0 && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.5, delay: 0.5 }}
            className="flex flex-wrap items-center gap-6 pt-4 border-t"
            style={{ borderColor: 'rgba(255,255,255,0.05)' }}
          >
            {[
              { color: '#3b82f6', label: 'Genre Match — based on your watch history' },
              { color: '#a855f7', label: 'AI Pick — suggested by Gemini 2.0' },
              { color: '#e50914', label: 'Top Pick — both engines agreed' },
            ].map(({ color, label }) => (
              <div key={label} className="flex items-center gap-2">
                <div
                  className="w-2 h-2 rounded-full flex-shrink-0"
                  style={{ backgroundColor: color }}
                />
                <span className="text-xs" style={{ color: '#4b5563' }}>
                  {label}
                </span>
              </div>
            ))}
          </motion.div>
        )}

      </div>
    </div>
  );
};

export default RecommendationPage;