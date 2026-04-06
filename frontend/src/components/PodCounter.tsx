import React from 'react';

interface PodCounterProps {
  name: string;
  current: number;
  max: number;
  color: string;
}

const PodCounter: React.FC<PodCounterProps> = ({ name, current, max, color }) => {
  const percentage = max > 0 ? (current / max) * 100 : 0;

  return (
    <div className="mb-5">
      <div className="flex justify-between items-center mb-2">
        <span className="text-sm font-medium text-gray-300">{name}</span>
        <span className="text-sm font-bold tabular-nums" style={{ color }}>
          {current} <span className="text-gray-500">/ {max}</span>
        </span>
      </div>
      <div className="w-full h-3 rounded-full bg-white/5 overflow-hidden">
        <div
          className="h-full rounded-full transition-all duration-700 ease-out"
          style={{
            width: `${percentage}%`,
            background: `linear-gradient(90deg, ${color}, ${color}99)`,
            boxShadow: `0 0 12px ${color}60`,
          }}
        />
      </div>
    </div>
  );
};

export default PodCounter;
