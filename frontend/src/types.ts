export type LoadType = 'cpu' | 'memory' | 'network';

export interface LoadState {
  active: boolean;
  intensity: number;
  loading: boolean;
}

export type LoadStates = {
  [key in LoadType]: LoadState;
};

export interface PodCount {
  current: number;
  desired: number;
  max: number;
}

export type PodCounts = {
  [key: string]: PodCount;
};

export interface ClusterMetrics {
  cpuUsagePercent: number;
  memoryUsagePercent: number;
}

export interface StreamEvent {
  time: string;
  message: string;
  type: 'info' | 'warning' | 'error' | 'success';
}

export interface SSEData {
  podCounts: PodCounts;
  loadStates: { [key in LoadType]: { active: boolean; intensity: number } };
  clusterMetrics: ClusterMetrics;
  events: StreamEvent[];
}

export interface ClusterInfo {
  nodeCount: number;
  kubernetesVersion: string;
}

export interface AuthState {
  isAdmin: boolean;
  authToken: string | null;
}

export interface LoginResponse {
  token: string;
  role: string;
}

export interface LoginError {
  error: string;
}
