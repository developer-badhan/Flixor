import React from 'react';
import { Cpu, Sparkles, Star } from 'lucide-react';
import type { RecommendSource } from '../types/recommendation.types';

interface SourceBadgeProps {
  source: RecommendSource;
  size?: 'sm' | 'md';
}

// ─── Badge configs per source ─────────────────────────────────────────────────

const BADGE_CONFIG = {
  rule: {
    label: 'Genre Match',
    icon: <Cpu size={10} />,
    style: {
      background: 'rgba(59,130,246,0.15)',
      color: '#60a5fa',
      border: '1px solid rgba(59,130,246,0.25)',
    },
  },
  ai: {
    label: 'AI Pick',
    icon: <Sparkles size={10} />,
    style: {
      background: 'rgba(168,85,247,0.15)',
      color: '#c084fc',
      border: '1px solid rgba(168,85,247,0.25)',
    },
  },
  'rule+ai': {
    label: 'Top Pick',
    icon: <Star size={10} fill="currentColor" />,
    style: {
      background: 'rgba(229,9,20,0.18)',
      color: '#f87171',
      border: '1px solid rgba(229,9,20,0.35)',
    },
  },
} as const;

const SourceBadge: React.FC<SourceBadgeProps> = ({ source, size = 'sm' }) => {
  const config = BADGE_CONFIG[source] ?? BADGE_CONFIG['rule'];
  const padX   = size === 'md' ? '8px' : '6px';
  const padY   = size === 'md' ? '4px' : '3px';
  const fs     = size === 'md' ? '11px' : '10px';

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        padding: `${padY} ${padX}`,
        borderRadius: '6px',
        fontSize: fs,
        fontWeight: 700,
        letterSpacing: '0.03em',
        whiteSpace: 'nowrap',
        ...config.style,
      }}
    >
      {config.icon}
      {config.label}
    </span>
  );
};

export default SourceBadge;