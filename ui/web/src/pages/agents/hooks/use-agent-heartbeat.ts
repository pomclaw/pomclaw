// Stub for removed heartbeat features
export interface HeartbeatConfig {
  [key: string]: any;
}

export interface DeliveryTarget {
  [key: string]: any;
}

export interface UseAgentHeartbeatReturn {
  config: null;
  loading: boolean;
  update: () => Promise<void>;
  test: () => Promise<void>;
  saving: boolean;
  getChecklist: () => any;
  setChecklist: (v: any) => void;
  fetchTargets: () => Promise<void>;
  refresh: () => void;
}

export function useAgentHeartbeat(_agentId: string): UseAgentHeartbeatReturn {
  return {
    config: null,
    loading: false,
    update: () => Promise.resolve(),
    test: () => Promise.resolve(),
    saving: false,
    getChecklist: () => [],
    setChecklist: () => {},
    fetchTargets: () => Promise.resolve(),
    refresh: () => {}
  };
}
