import React from 'react';
import { motion } from 'framer-motion';
import { Cpu, Sparkles, Layers } from 'lucide-react';
import type { RecommendMode } from '../types/recommendation.types';

// ─── Tab config ───────────────────────────────────────────────────────────────

interface ModeTab {
  value: RecommendMode;
  label: string;
  sublabel: string;
  icon: React.ReactNode;
  accentColor: string;
}

const TABS: ModeTab[] = [
  {
    value: 'hybrid',
    label: 'Hybrid',
    sublabel: 'Best of both',
    icon: <Layers size={16} />,
    accentColor: '#e50914',
  },
  {
    value: 'rule',
    label: 'Genre Match',
    sublabel: 'Based on history',
    icon: <Cpu size={16} />,
    accentColor: '#3b82f6',
  },
  {
    value: 'ai',
    label: 'AI Powered',
    sublabel: 'Gemini 2.0',
    icon: <Sparkles size={16} />,
    accentColor: '#a855f7',
  },
];

// ─── Props ────────────────────────────────────────────────────────────────────

interface ModeSelectorProps {
  activeMode: RecommendMode;
  loading: boolean;
  onChange: (mode: RecommendMode) => void;
}

// ─── Component ────────────────────────────────────────────────────────────────

const ModeSelector: React.FC<ModeSelectorProps> = ({
  activeMode,
  loading,
  onChange,
}) => {
  return (
    <div className="flex flex-col sm:flex-row gap-3">
      {TABS.map((tab) => {
        const isActive = activeMode === tab.value;
        return (
          <button
            key={tab.value}
            onClick={() => !loading && onChange(tab.value)}
            disabled={loading}
            className="relative flex items-center gap-3 px-5 py-3.5 rounded-xl border transition-all duration-250 text-left flex-1"
            style={{
              background: isActive
                ? `${tab.accentColor}12`
                : 'rgba(255,255,255,0.03)',
              borderColor: isActive
                ? `${tab.accentColor}40`
                : 'rgba(255,255,255,0.07)',
              cursor: loading ? 'not-allowed' : 'pointer',
              opacity: loading && !isActive ? 0.5 : 1,
            }}
          >
            {/* Active glow */}
            {isActive && (
              <motion.div
                layoutId="mode-glow"
                className="absolute inset-0 rounded-xl pointer-events-none"
                style={{
                  background: `radial-gradient(ellipse at left center, ${tab.accentColor}10, transparent 70%)`,
                }}
                transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              />
            )}

            {/* Icon circle */}
            <div
              className="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors duration-200"
              style={{
                backgroundColor: isActive
                  ? `${tab.accentColor}20`
                  : 'rgba(255,255,255,0.05)',
                color: isActive ? tab.accentColor : '#6b7280',
              }}
            >
              {tab.icon}
            </div>

            {/* Text */}
            <div className="min-w-0">
              <p
                className="text-sm font-bold transition-colors duration-200 leading-tight"
                style={{ color: isActive ? '#fff' : '#9ca3af' }}
              >
                {tab.label}
              </p>
              <p
                className="text-xs mt-0.5 transition-colors duration-200"
                style={{ color: isActive ? tab.accentColor : '#4b5563' }}
              >
                {tab.sublabel}
              </p>
            </div>

            {/* Active indicator dot */}
            {isActive && (
              <motion.div
                layoutId="mode-dot"
                className="absolute top-3 right-3 w-1.5 h-1.5 rounded-full"
                style={{ backgroundColor: tab.accentColor }}
                transition={{ type: 'spring', stiffness: 400, damping: 30 }}
              />
            )}
          </button>
        );
      })}
    </div>
  );
};

export default ModeSelector;