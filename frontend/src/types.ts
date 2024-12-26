export type LoadType = 'cpu' | 'memory' | 'network';

export interface LoadState {
  loading: boolean;
  active: boolean;
}

export type LoadStates = {
  [key in LoadType]: LoadState;
}