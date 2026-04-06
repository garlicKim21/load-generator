import React from 'react';
import { Activity } from 'lucide-react';
import CircularGauge from './CircularGauge';
import PodCounter from './PodCounter';
import { PodCounts, ClusterMetrics } from '../types';

interface MetricsPanelProps {
  podCounts: PodCounts;
  clusterMetrics: ClusterMetrics;
}

const podColorMap: Record<string, string> = {
  'cpu-generator': '#60a5fa',
  'memory-generator': '#c084fc',
  'network-generator': '#34d399',
};

const MetricsPanel: React.FC<MetricsPanelProps> = ({ podCounts, clusterMetrics }) => {
  return (
    <div className="h-full flex flex-col gap-6">
      {/* Cluster gauges */}
      <div className="bg-white/5 backdrop-blur-lg border border-white/10 rounded-2xl p-6">
        <div className="flex items-center gap-2 mb-6">
          <Activity size={18} className="text-gray-400" />
          <h2 className="text-base font-semibold text-white">Cluster Utilization</h2>
        </div>
        <div className="flex justify-around">
          <CircularGauge
            percentage={clusterMetrics.cpuUsagePercent}
            label="CPU"
            color="#60a5fa"
          />
          <CircularGauge
            percentage={clusterMetrics.memoryUsagePercent}
            label="Memory"
            color="#c084fc"
          />
        </div>
      </div>

      {/* Pod counts */}
      <div className="bg-white/5 backdrop-blur-lg border border-white/10 rounded-2xl p-6 flex-1">
        <div className="flex items-center gap-2 mb-5">
          <Activity size={18} className="text-gray-400" />
          <h2 className="text-base font-semibold text-white">Pod Scaling</h2>
        </div>
        {Object.entries(podCounts).length === 0 ? (
          <p className="text-gray-500 text-sm text-center mt-4">
            No pod data available yet.
          </p>
        ) : (
          Object.entries(podCounts).map(([name, counts]) => (
            <PodCounter
              key={name}
              name={name}
              current={counts.current}
              max={counts.max}
              color={podColorMap[name] || '#60a5fa'}
            />
          ))
        )}
      </div>
    </div>
  );
};

export default MetricsPanel;
