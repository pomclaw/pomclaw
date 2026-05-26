export const queryKeys = {
  traces: {
    all: ["traces"] as const,
    list: (params: Record<string, unknown>) => ["traces", params] as const,
  },
  // Add other query keys as needed
};
