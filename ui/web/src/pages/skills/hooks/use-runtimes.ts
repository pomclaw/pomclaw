// Stub for removed skills page
export interface RuntimeInfo {
  name: string;
  version?: string;
  available: boolean;
}

export function useRuntimes() {
  return {
    runtimes: { runtimes: [] as RuntimeInfo[] },
    loading: false
  };
}
