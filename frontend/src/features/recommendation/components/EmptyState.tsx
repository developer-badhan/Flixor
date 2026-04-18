import React from 'react';
import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Film, TrendingUp, Sparkles } from 'lucide-react';

interface EmptyStateProps {
  message?: string;
}

// ─── Floating icon animation variants ────────────────────────────────────────

const floatVariants = {
  animate: (delay: number) => ({
    y: [0, -8, 0],
    transition: {
      duration: 3,
      repeat: Infinity,
      repeatType: 'loop' as const,
      ease: 'easeInOut',
      delay,
    },
  }),
};

const EmptyState: React.FC<EmptyStateProps> = ({ message }) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: 'easeOut' }}
      className="flex flex-col items-center justify-center py-24 px-6 text-center"
    >
      {/* Illustration — 3 floating icons in a cluster */}
      <div className="relative w-36 h-36 mb-10">
        {/* Outer glow ring */}
        <div
          className="absolute inset-0 rounded-full opacity-20"
          style={{
            background:
              'radial-gradient(circle, rgba(229,9,20,0.4) 0%, transparent 70%)',
          }}
        />

        {/* Center icon */}
        <motion.div
          custom={0}
          variants={floatVariants}
          animate="animate"
          className="absolute inset-0 flex items-center justify-center"
        >
          <div
            className="w-16 h-16 rounded-2xl flex items-center justify-center"
            style={{ background: 'rgba(229,9,20,0.12)', border: '1px solid rgba(229,9,20,0.2)' }}
          >
            <Film size={30} style={{ color: '#e50914' }} />
          </div>
        </motion.div>

        {/* Top-right floating icon */}
        <motion.div
          custom={0.8}
          variants={floatVariants}
          animate="animate"
          className="absolute -top-1 right-2"
        >
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center"
            style={{ background: 'rgba(168,85,247,0.12)', border: '1px solid rgba(168,85,247,0.2)' }}
          >
            <Sparkles size={18} style={{ color: '#a855f7' }} />
          </div>
        </motion.div>

        {/* Bottom-left floating icon */}
        <motion.div
          custom={1.6}
          variants={floatVariants}
          animate="animate"
          className="absolute -bottom-1 left-2"
        >
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center"
            style={{ background: 'rgba(59,130,246,0.12)', border: '1px solid rgba(59,130,246,0.2)' }}
          >
            <TrendingUp size={18} style={{ color: '#3b82f6' }} />
          </div>
        </motion.div>
      </div>

      {/* Heading */}
      <h3 className="text-2xl font-bold text-white mb-3 tracking-tight">
        No Recommendations Yet
      </h3>

      {/* Backend message or fallback */}
      <p
        className="text-sm max-w-sm leading-relaxed mb-8"
        style={{ color: '#6b7280' }}
      >
        {message ||
          "Watch a few movies to get personalised picks! Our engines analyse your genre preferences and use AI to find your perfect match."}
      </p>

      {/* How it works — 3 small steps */}
      <div
        className="flex flex-col sm:flex-row gap-4 mb-10 w-full max-w-md"
      >
        {[
          { icon: <Film size={14} />, label: 'Watch movies', color: '#e50914' },
          { icon: <TrendingUp size={14} />, label: 'Build your taste profile', color: '#3b82f6' },
          { icon: <Sparkles size={14} />, label: 'Get AI recommendations', color: '#a855f7' },
        ].map((step, i) => (
          <div
            key={i}
            className="flex items-center gap-2.5 flex-1 px-4 py-3 rounded-xl"
            style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.06)' }}
          >
            <div
              className="w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0"
              style={{ backgroundColor: `${step.color}18`, color: step.color }}
            >
              {step.icon}
            </div>
            <span className="text-xs font-medium" style={{ color: '#9ca3af' }}>
              {step.label}
            </span>
          </div>
        ))}
      </div>

      {/* CTA */}
      <div className="flex items-center gap-3">
        <Link
          to="/movies"
          className="flex items-center gap-2 px-6 py-3 rounded-xl font-bold text-sm text-white transition-all duration-200 hover:opacity-90 active:scale-95"
          style={{ background: '#e50914' }}
        >
          <Film size={16} />
          Browse Movies
        </Link>
        <Link
          to="/analytics"
          className="flex items-center gap-2 px-6 py-3 rounded-xl font-bold text-sm transition-all duration-200 hover:bg-white/5"
          style={{
            color: '#9ca3af',
            border: '1px solid rgba(255,255,255,0.08)',
          }}
        >
          <TrendingUp size={16} />
          View Trending
        </Link>
      </div>
    </motion.div>
  );
};

export default EmptyState;