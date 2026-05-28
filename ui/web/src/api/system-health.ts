import { useQuery } from "@tanstack/react-query";
import { useHttp } from "@/hooks/use-ws";

export interface SystemHealth {
  version?: string;
  uptime?: number;
  database?: string;
  tools?: number;
  sessions?: number;
  providers?: number;
  channelTotal?: number;
  channelOnline?: number;
  channelDegraded?: number;
  channelFailed?: number;
}

export interface SystemHealthResp {
  health: SystemHealth;
}

export const systemHealthKeys = {
  all: ["system-health"] as const,
};

export function useSystemHealth() {
  const http = useHttp();
  return useQuery({
    queryKey: systemHealthKeys.all,
    queryFn: async () => {
      const resp = await http.get<SystemHealthResp>("/v1/system/health");
      return resp.health;
    },
    refetchInterval: 30_000,
    retry: 1,
  });
}
