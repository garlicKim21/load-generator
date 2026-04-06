import React from 'react';
import { Cpu, MemoryStick, Wifi, Loader2, Lock } from 'lucide-react';
import { LoadType, LoadState } from '../types';

interface LoadControlCardProps {
  type: LoadType;
  state: LoadState;
  onToggle: () => void;
  onIntensityChange: (level: number) => void;
  isAdmin?: boolean;
  authToken?: string | null;
}

const config: Record<LoadType, { icon: React.FC<any>; label: string; color: string; accent: string }> = {
  cpu: { icon: Cpu, label: 'CPU Load', color: '#60a5fa', accent: 'blue' },
  memory: { icon: MemoryStick, label: 'Memory Load', color: '#c084fc', accent: 'purple' },
  network: { icon: Wifi, label: 'Network Load', color: '#34d399', accent: 'green' },
};

const LoadControlCard: React.FC<LoadControlCardProps> = ({
  type,
  state,
  onToggle,
  onIntensityChange,
  isAdmin = false,
  authToken,
}) => {
  const { icon: Icon, label, color } = config[type];
  const { active, intensity, loading } = state;
  const disabled = !isAdmin;

  return (
    <div
      className="relative bg-white/5 backdrop-blur-lg border border-white/10 rounded-2xl p-6 transition-all duration-300 hover:bg-white/8"
      style={{
        borderColor: active ? `${color}40` : undefined,
        boxShadow: active ? `0 0 30px ${color}15, inset 0 0 30px ${color}05` : undefined,
      }}
    >
      {/* Top row: icon + title + status badge */}
      <div className="flex items-center justify-between mb-5">
        <div className="flex items-center gap-3">
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center"
            style={{ backgroundColor: `${color}20` }}
          >
            <Icon size={20} style={{ color }} />
          </div>
          <h3 className="text-lg font-semibold text-white">{label}</h3>
        </div>
        <span
          className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider"
          style={{
            color: active ? color : '#9ca3af',
            backgroundColor: active ? `${color}20` : 'rgba(255,255,255,0.05)',
          }}
        >
          {active ? 'Active' : 'Idle'}
        </span>
      </div>

      {/* Intensity slider */}
      <div className="mb-5">
        <div className="flex justify-between items-center mb-2">
          <span className="text-sm text-gray-400">Intensity</span>
          <span className="text-sm font-bold tabular-nums" style={{ color: active ? color : '#9ca3af' }}>
            {intensity} / 10
          </span>
        </div>
        <input
          type="range"
          min={1}
          max={10}
          value={intensity}
          onChange={(e) => onIntensityChange(Number(e.target.value))}
          disabled={disabled}
          className={`w-full h-2 rounded-full appearance-none slider ${disabled ? 'cursor-not-allowed opacity-40' : 'cursor-pointer'}`}
          style={{
            background: `linear-gradient(to right, ${color} 0%, ${color} ${((intensity - 1) / 9) * 100}%, rgba(255,255,255,0.1) ${((intensity - 1) / 9) * 100}%, rgba(255,255,255,0.1) 100%)`,
          }}
        />
        {/* Tick marks */}
        <div className="flex justify-between mt-1 px-0.5">
          {Array.from({ length: 10 }, (_, i) => (
            <span key={i} className="text-[10px] text-gray-600 tabular-nums">{i + 1}</span>
          ))}
        </div>
      </div>

      {/* Toggle button */}
      <button
        onClick={onToggle}
        disabled={loading || disabled}
        className="w-full py-3 rounded-xl font-semibold text-sm uppercase tracking-wider transition-all duration-300 flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        style={{
          backgroundColor: disabled ? 'rgba(255,255,255,0.05)' : active ? `${color}20` : `${color}`,
          color: disabled ? '#6b7280' : active ? color : '#0f172a',
          border: active && !disabled ? `1px solid ${color}50` : 'none',
        }}
      >
        {loading ? (
          <>
            <Loader2 size={16} className="animate-spin" />
            Processing...
          </>
        ) : disabled ? (
          <span className="flex items-center gap-1.5 text-gray-500">
            <Lock size={14} />
            Admin only
          </span>
        ) : active ? (
          'Stop Generator'
        ) : (
          'Start Generator'
        )}
      </button>
    </div>
  );
};

export default LoadControlCard;
