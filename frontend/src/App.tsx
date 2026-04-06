import React, { useState, useEffect, useCallback, useRef } from 'react';
import axios from 'axios';
import { Boxes, Wifi, WifiOff, Server, Code, Lock, Unlock } from 'lucide-react';
import LoadControlCard from './components/LoadControlCard';
import MetricsPanel from './components/MetricsPanel';
import EventFeed from './components/EventFeed';
import LoginModal from './components/LoginModal';
import {
  LoadType,
  LoadStates,
  PodCounts,
  ClusterMetrics,
  StreamEvent,
  SSEData,
  ClusterInfo,
} from './types';

const initialLoadStates: LoadStates = {
  cpu: { active: false, intensity: 5, loading: false },
  memory: { active: false, intensity: 5, loading: false },
  network: { active: false, intensity: 5, loading: false },
};

const MAX_EVENTS = 50;

function App() {
  const [loadStates, setLoadStates] = useState<LoadStates>(initialLoadStates);
  const [podCounts, setPodCounts] = useState<PodCounts>({});
  const [clusterMetrics, setClusterMetrics] = useState<ClusterMetrics>({
    cpuUsagePercent: 0,
    memoryUsagePercent: 0,
  });
  const [events, setEvents] = useState<StreamEvent[]>([]);
  const [connected, setConnected] = useState(false);
  const [clusterInfo, setClusterInfo] = useState<ClusterInfo>({
    nodeCount: 0,
    kubernetesVersion: '',
  });
  const [isAdmin, setIsAdmin] = useState(false);
  const [authToken, setAuthToken] = useState<string | null>(null);
  const [showLoginModal, setShowLoginModal] = useState(false);

  const eventSourceRef = useRef<EventSource | null>(null);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  // Validate stored token on mount
  useEffect(() => {
    const storedToken = localStorage.getItem('auth_token');
    if (storedToken) {
      axios
        .get('/api/v1/auth/validate', {
          headers: { Authorization: `Bearer ${storedToken}` },
        })
        .then(() => {
          setAuthToken(storedToken);
          setIsAdmin(true);
        })
        .catch(() => {
          localStorage.removeItem('auth_token');
        });
    }
  }, []);

  // Add an event, keeping max 50
  const addEvents = useCallback((newEvents: StreamEvent[]) => {
    setEvents((prev) => {
      const combined = [...prev, ...newEvents];
      return combined.slice(-MAX_EVENTS);
    });
  }, []);

  // Apply SSE data to state
  const applySSEData = useCallback(
    (data: SSEData) => {
      if (data.podCounts) setPodCounts(data.podCounts);
      if (data.clusterMetrics) setClusterMetrics(data.clusterMetrics);
      if (data.loadStates) {
        setLoadStates((prev) => {
          const next = { ...prev };
          for (const key of Object.keys(data.loadStates) as LoadType[]) {
            next[key] = {
              ...next[key],
              active: data.loadStates[key].active,
              intensity: data.loadStates[key].intensity,
            };
          }
          return next;
        });
      }
      if (data.events && data.events.length > 0) {
        addEvents(data.events);
      }
    },
    [addEvents]
  );

  // Polling-based data fetching
  useEffect(() => {
    let active = true;

    const fetchData = async () => {
      try {
        const [metricsRes, statusRes] = await Promise.all([
          axios.get('/api/v1/metrics'),
          axios.get('/api/v1/status'),
        ]);

        if (!active) return;
        setConnected(true);

        // Apply metrics (podCounts, clusterMetrics)
        if (metricsRes.data) {
          applySSEData(metricsRes.data);
        }

        // Apply status (load states)
        const statusData = statusRes.data;
        if (statusData) {
          const mapped: { [key: string]: { active: boolean; intensity: number } } = {};
          for (const t of ['cpu', 'memory', 'network'] as LoadType[]) {
            const s = statusData[t] || statusData[t.toUpperCase()];
            if (s) {
              mapped[t] = {
                active: s.state === 'active',
                intensity: s.intensity ?? 5,
              };
            }
          }
          if (Object.keys(mapped).length > 0) {
            setLoadStates((prev) => {
              const next = { ...prev };
              for (const key of Object.keys(mapped) as LoadType[]) {
                next[key] = { ...next[key], ...mapped[key] };
              }
              return next;
            });
          }
        }
      } catch {
        if (active) setConnected(false);
      }
    };

    fetchData();
    pollIntervalRef.current = setInterval(fetchData, 2000);

    return () => {
      active = false;
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
    };
  }, [applySSEData]);

  // Auth helpers
  const getAuthHeaders = useCallback(() => {
    if (authToken) {
      return { Authorization: `Bearer ${authToken}` };
    }
    return {};
  }, [authToken]);

  const handleLoginSuccess = useCallback((token: string) => {
    setAuthToken(token);
    setIsAdmin(true);
    setShowLoginModal(false);
  }, []);

  const handleLogout = useCallback(() => {
    localStorage.removeItem('auth_token');
    setAuthToken(null);
    setIsAdmin(false);
  }, []);

  // Toggle load generator
  const handleToggle = useCallback(
    async (type: LoadType) => {
      setLoadStates((prev) => ({
        ...prev,
        [type]: { ...prev[type], loading: true },
      }));

      const action = loadStates[type].active ? 'stop' : 'start';
      try {
        await axios.post(`/api/v1/load/${type}/${action}`, {}, {
          headers: getAuthHeaders(),
        });
        setLoadStates((prev) => ({
          ...prev,
          [type]: { ...prev[type], active: !prev[type].active, loading: false },
        }));
        addEvents([
          {
            time: new Date().toLocaleTimeString('en-US', { hour12: false }),
            message: `${type.toUpperCase()} load ${action === 'start' ? 'started' : 'stopped'} (intensity ${loadStates[type].intensity})`,
            type: action === 'start' ? 'success' : 'info',
          },
        ]);
      } catch (err: any) {
        setLoadStates((prev) => ({
          ...prev,
          [type]: { ...prev[type], loading: false },
        }));
        addEvents([
          {
            time: new Date().toLocaleTimeString('en-US', { hour12: false }),
            message: `Failed to ${action} ${type.toUpperCase()} load: ${err.message}`,
            type: 'error',
          },
        ]);
      }
    },
    [loadStates, addEvents, getAuthHeaders]
  );

  // Change intensity
  const handleIntensityChange = useCallback(
    async (type: LoadType, level: number) => {
      // Optimistic update
      setLoadStates((prev) => ({
        ...prev,
        [type]: { ...prev[type], intensity: level },
      }));

      try {
        await axios.post(`/api/v1/load/${type}/intensity`, { level }, {
          headers: getAuthHeaders(),
        });
      } catch (err) {
        console.error('Failed to set intensity:', err);
      }
    },
    [getAuthHeaders]
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 text-white flex flex-col">
      {/* Header */}
      <header className="flex items-center justify-between px-8 py-4 border-b border-white/10 bg-white/5 backdrop-blur-lg">
        <div className="flex items-center gap-3">
          <Boxes size={28} className="text-blue-400" />
          <h1 className="text-2xl font-bold tracking-tight">
            Kubernetes Load Generator
          </h1>
        </div>
        <div className="flex items-center gap-4">
          {connected ? (
            <div className="flex items-center gap-2">
              <span className="relative flex h-3 w-3">
                <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                <span className="relative inline-flex rounded-full h-3 w-3 bg-green-400"></span>
              </span>
              <Wifi size={16} className="text-green-400" />
              <span className="text-sm text-green-400 font-medium">Connected</span>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <span className="h-3 w-3 rounded-full bg-red-400"></span>
              <WifiOff size={16} className="text-red-400" />
              <span className="text-sm text-red-400 font-medium">Disconnected</span>
            </div>
          )}

          {/* Auth button */}
          <div className="border-l border-white/10 pl-4">
            {isAdmin ? (
              <button
                onClick={handleLogout}
                className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-green-500/10 border border-green-500/20 hover:bg-green-500/20 transition-colors"
              >
                <Unlock size={14} className="text-green-400" />
                <span className="text-sm font-medium text-green-400">Admin</span>
              </button>
            ) : (
              <button
                onClick={() => setShowLoginModal(true)}
                className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 hover:bg-white/10 transition-colors"
              >
                <Lock size={14} className="text-gray-400" />
                <span className="text-sm font-medium text-gray-400">Login</span>
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Main content: 3-column layout */}
      <main className="flex-1 grid grid-cols-1 lg:grid-cols-[1fr_1.2fr_1fr] gap-6 p-6 overflow-hidden">
        {/* Left: Load Controls */}
        <div className="flex flex-col gap-5">
          <h2 className="text-xs font-bold uppercase tracking-widest text-gray-500 pl-1">
            Load Controls
          </h2>
          {(['cpu', 'memory', 'network'] as LoadType[]).map((type) => (
            <LoadControlCard
              key={type}
              type={type}
              state={loadStates[type]}
              onToggle={() => handleToggle(type)}
              onIntensityChange={(level) => handleIntensityChange(type, level)}
              isAdmin={isAdmin}
              authToken={authToken}
            />
          ))}
        </div>

        {/* Center: Metrics */}
        <div>
          <h2 className="text-xs font-bold uppercase tracking-widest text-gray-500 pl-1 mb-5">
            Real-time Metrics
          </h2>
          <MetricsPanel podCounts={podCounts} clusterMetrics={clusterMetrics} />
        </div>

        {/* Right: Event Feed */}
        <div className="flex flex-col min-h-0">
          <h2 className="text-xs font-bold uppercase tracking-widest text-gray-500 pl-1 mb-5">
            Activity Log
          </h2>
          <div className="flex-1 min-h-0">
            <EventFeed events={events} />
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="flex items-center justify-center gap-6 px-8 py-3 border-t border-white/10 bg-white/5 text-sm text-gray-500">
        <div className="flex items-center gap-2">
          <Server size={14} />
          <span>Nodes: {clusterInfo.nodeCount || '...'}</span>
        </div>
        <div className="flex items-center gap-2">
          <Code size={14} />
          <span>K8s: {clusterInfo.kubernetesVersion || '...'}</span>
        </div>
      </footer>

      {/* Login Modal */}
      {showLoginModal && (
        <LoginModal
          onSuccess={handleLoginSuccess}
          onClose={() => setShowLoginModal(false)}
        />
      )}
    </div>
  );
}

export default App;
