import React from 'react';
import { motion } from 'framer-motion';
import { BarChart2 } from 'lucide-react';
import { StatCards, TrendingMovies, MostWatched, TopGenres } from '../features/analytics';

// ─── Section fade-in wrapper ──────────────────────────────────────────────────

const Section: React.FC<{ children: React.ReactNode; delay?: number }> = ({ children, delay = 0 }) => (
  <motion.section
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    transition={{ duration: 0.5, delay, ease: 'easeOut' }}
  >
    {children}
  </motion.section>
);

// ─── Page ─────────────────────────────────────────────────────────────────────

const AnalyticsPage: React.FC = () => {
  return (
    <div className="min-h-screen bg-flixor-dark pt-24 pb-20">
      {/* Subtle noise texture overlay for depth */}
      <div
        className="fixed inset-0 pointer-events-none z-0 opacity-[0.025]"
        style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg viewBox='0 0 256 256' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='4' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E")`,
        }}
      />

      <div className="relative z-10 max-w-7xl mx-auto px-6 md:px-10 space-y-10">

        {/* ── Page Header ── */}
        <Section delay={0}>
          <div className="flex items-center gap-4 mb-2">
            <div className="w-10 h-10 rounded-xl bg-flixor-red/15 flex items-center justify-center text-flixor-red">
              <BarChart2 size={22} />
            </div>
            <div>
              <h1 className="text-3xl md:text-4xl font-bold text-white tracking-tight">
                Platform Analytics
              </h1>
              <p className="text-flixor-lightGray text-sm mt-0.5">
                Real-time insights across movies, users, and engagement
              </p>
            </div>
          </div>

          {/* Decorative separator */}
          <div className="mt-6 h-px bg-gradient-to-r from-flixor-red/40 via-white/10 to-transparent" />
        </Section>

        {/* ── KPI Stat Cards ── */}
        <Section delay={0.1}>
          <StatCards />
        </Section>

        {/* ── Trending Movies (full-width) ── */}
        <Section delay={0.2}>
          <TrendingMovies />
        </Section>

        {/* ── Most Watched + Top Genres (side-by-side) ── */}
        <Section delay={0.3}>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <MostWatched />
            <TopGenres />
          </div>
        </Section>

      </div>
    </div>
  );
};

export default AnalyticsPage;