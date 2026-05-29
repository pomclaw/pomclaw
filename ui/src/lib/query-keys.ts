export const queryKeys = {
  traces: {
    all: ["traces"] as const,
    list: (params: Record<string, unknown>) => ["traces", params] as const,
  },
  memory: {
    all: ["memory"] as const,
    list: (params: Record<string, unknown>) => ["memory", params] as const,
  },
  episodic: {
    all: ["episodic"] as const,
    list: (agentId: string, params: Record<string, unknown>) => ["episodic", agentId, params] as const,
  },
  agents: {
    all: ["agents"] as const,
    detail: (id: string) => ["agents", id] as const,
  },
  contacts: {
    all: ["contacts"] as const,
    resolve: (ids: string) => ["contacts", "resolve", ids] as const,
  },
  // Add other query keys as needed
};
