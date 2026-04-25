import React from 'react';
import { motion } from 'framer-motion';
import { Film, Users, Eye, Bookmark, ThumbsUp } from 'lucide-react';
import { usePlatformStats } from '../../../hooks/useAnalytics';

/**
 * Formats a large number with appropriate suffixes (K, M, etc.).
 * @param n - The number to format.
 * @returns The formatted number string.
 * @function formatCount
 * @returns {string}
 */
function formatCount(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
  return n.toString();
}

/**
 * Interface for the stat card configuration.
 * @interface StatConfig
 * @property {keyof ReturnType<typeof usePlatformStats>['data'] & string} key - The key of the stat.
 * @property {string} label - The label of the stat.
 * @property {React.ReactNode} icon - The icon of the stat.
 * @property {string} accentColor - The accent color of the stat.
 * @property {string} glowColor - The glow color of the stat.
 */
interface StatConfig {
  key: keyof ReturnType<typeof usePlatformStats>['data'] & string;
  label: string;
  icon: React.ReactNode;
  accentColor: string;
  glowColor: string;
}

const STAT_CARDS: StatConfig[] = [
  {
    key: 'total_movies',
    label: 'Total Movies',
    icon: <Film size={22} />,
    accentColor: '#e50914',
    glowColor: 'rgba(229,9,20,0.15)',
  },
  {
    key: 'total_users',
    label: 'Total Users',
    icon: <Users size={22} />,
    accentColor: '#3b82f6',
    glowColor: 'rgba(59,130,246,0.15)',
  },
  {
    key: 'total_views',
    label: 'Total Views',
    icon: <Eye size={22} />,
    accentColor: '#10b981',
    glowColor: 'rgba(16,185,129,0.15)',
  },
  {
    key: 'total_watchlist',
    label: 'Watchlist Entries',
    icon: <Bookmark size={22} />,
    accentColor: '#f59e0b',
    glowColor: 'rgba(245,158,11,0.15)',
  },
  {
    key: 'total_likes',
    label: 'Total Likes',
    icon: <ThumbsUp size={22} />,
    accentColor: '#8b5cf6',
    glowColor: 'rgba(139,92,246,0.15)',
  },
];

// ─── Skeleton ─────────────────────────────────────────────────────────────────

const CardSkeleton: React.FC = () => (
  <div className="rounded-xl bg-flixor-gray/40 border border-white/5 p-6 animate-pulse">
    <div className="flex items-center justify-between mb-4">
      <div className="w-10 h-10 rounded-lg bg-flixor-gray" />
      <div className="w-16 h-3 rounded bg-flixor-gray" />
    </div>
    <div className="w-24 h-8 rounded bg-flixor-gray mb-1" />
    <div className="w-20 h-3 rounded bg-flixor-gray/60" />
  </div>
);

// ─── Single Card ──────────────────────────────────────────────────────────────

interface StatCardProps {
  config: StatConfig;
  value: number;
  index: number;
}

const StatCard: React.FC<StatCardProps> = ({ config, value, index }) => (
  <motion.div
    initial={{ opacity: 0, y: 24 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.45, delay: index * 0.08, ease: 'easeOut' }}
    className="relative rounded-xl border border-white/8 p-6 overflow-hidden group"
    style={{
      background: `linear-gradient(135deg, #1c1c1c 0%, #181818 100%)`,
      boxShadow: `0 0 0 1px rgba(255,255,255,0.04) inset`,
    }}
    whileHover={{ y: -3, transition: { duration: 0.2 } }}
  >
    {/* Glow on hover */}
    <div
      className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 rounded-xl pointer-events-none"
      style={{ background: `radial-gradient(ellipse at top left, ${config.glowColor}, transparent 70%)` }}
    />

    {/* Top row: icon + accent bar */}
    <div className="flex items-center justify-between mb-5">
      <div
        className="w-10 h-10 rounded-lg flex items-center justify-center"
        style={{ backgroundColor: `${config.accentColor}20`, color: config.accentColor }}
      >
        {config.icon}
      </div>
      {/* Decorative mini accent line */}
      <div
        className="h-0.5 w-12 rounded-full opacity-40"
        style={{ backgroundColor: config.accentColor }}
      />
    </div>

    {/* Value */}
    <p
      className="text-3xl font-bold tracking-tight mb-1"
      style={{ fontVariantNumeric: 'tabular-nums', color: '#fff' }}
    >
      {formatCount(value)}
    </p>

    {/* Label */}
    <p className="text-xs text-flixor-lightGray uppercase tracking-widest font-medium">
      {config.label}
    </p>

    {/* Bottom accent border */}
    <div
      className="absolute bottom-0 left-0 right-0 h-0.5 opacity-0 group-hover:opacity-60 transition-opacity duration-300"
      style={{ backgroundColor: config.accentColor }}
    />
  </motion.div>
);

// ─── Main Component ───────────────────────────────────────────────────────────

const StatCards: React.FC = () => {
  const { data, loading, error } = usePlatformStats();

  if (error) {
    return (
      <div className="rounded-xl border border-flixor-red/30 bg-flixor-red/10 p-4 text-flixor-red text-sm">
        {error}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
      {loading
        ? STAT_CARDS.map((_, i) => <CardSkeleton key={i} />)
        : STAT_CARDS.map((config, i) => (
            <StatCard
              key={config.key}
              config={config}
              value={(data as any)?.[config.key] ?? 0}
              index={i}
            />
          ))}
    </div>
  );
};

export default StatCards;