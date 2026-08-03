/**
 * 系统健康检查 - 使用生成的 API
 * 替代 @/api/system-health.ts
 */

import { useQuery } from "@tanstack/react-query";
import { useApiClient } from "./use-api-client";

export interface SystemHealth {
  version?: string;
  uptime?: number;
  tools?: number;
  sessions?: number;
  providers?: number;
  skills?: number;
  memory?: number;
  document?: number;
  channelTotal?: number;
  channelOnline?: number;
  channelDegraded?: number;
  channelFailed?: number;
}

export interface QuotaUsage {
  requests_today?: number;
  input_tokens_today?: number;
  output_tokens_today?: number;
  cost_today?: number;
}

export interface SystemHealthResp {
  health: SystemHealth;
  quota_usage: QuotaUsage;
}

export const systemHealthKeys = {
  all: ["system-health"] as const,
};

/**
 * Hook: 获取系统健康状态和配额使用
 * 使用生成的 API (getSystemHealth) 通过 useApiClient
 */
export function useSystemHealth() {
  const api = useApiClient();

  return useQuery({
    queryKey: systemHealthKeys.all,
    queryFn: async () => {
      const resp = await api.getSystemHealth() as SystemHealthResp;
      return {
        health: resp.health,
        quotaUsage: resp.quota_usage,
      };
    },
    refetchInterval: 30_000,
    retry: 1,
  });
}
