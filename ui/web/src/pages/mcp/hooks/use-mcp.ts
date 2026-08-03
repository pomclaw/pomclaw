import { useCallback, useRef } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useApiClient } from "@/hooks/use-api-client";
import { queryKeys } from "@/lib/query-keys";
import { toast } from "@/stores/use-toast-store";
import i18n from "@/i18n";
import { userFriendlyError } from "@/lib/error-utils";
import type {
  MCPToolInfo,
  MCPAgentGrant,
  CreateMCPServerReq,
  UpdateMCPServerReq,
} from "@/client/pomclawComponents";

export function useMCP() {
  const api = useApiClient();
  const queryClient = useQueryClient();
  const refreshRef = useRef<(() => Promise<void>) | null>(null);

  // ── Read ──
  const { data: servers = [], isPending: loading, error: queryError } = useQuery({
    queryKey: queryKeys.mcp.all,
    queryFn: async () => {
      const res = await api.listMCPServers();
      return res.servers ?? [];
    },
    staleTime: 60_000,
  });

  const error = queryError instanceof Error ? queryError.message : queryError ? "Failed to load MCP servers" : null;

  const invalidate = useCallback(
    () => queryClient.invalidateQueries({ queryKey: queryKeys.mcp.all }),
    [queryClient],
  );

  // ── CRUD ──
  const createServer = useCallback(
    async (data: CreateMCPServerReq) => {
      try {
        const res = await api.createMCPServer(data);
        await invalidate();
        toast.success(i18n.t("mcp:toast.created"));
        return res.server;
      } catch (err) {
        toast.error(i18n.t("mcp:toast.failedCreate"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const updateServer = useCallback(
    async (id: string, data: UpdateMCPServerReq) => {
      try {
        const res = await api.updateMCPServer({}, data, id);
        await invalidate();
        toast.success(i18n.t("mcp:toast.updated"));
        return res.server;
      } catch (err) {
        toast.error(i18n.t("mcp:toast.failedUpdate"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const deleteServer = useCallback(
    async (id: string) => {
      try {
        await api.deleteMCPServer({}, id);
        await invalidate();
        toast.success(i18n.t("mcp:toast.deleted"));
      } catch (err) {
        toast.error(i18n.t("mcp:toast.failedDelete"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  // ── Connection ──
  const testConnection = useCallback(
    async (data: CreateMCPServerReq) => {
      const res = await api.testMCPServerConnection(data);
      return res;
    },
    [api],
  );

  const reconnectServer = useCallback(
    async (id: string) => {
      try {
        await api.reconnectMCPServer({}, id);
        toast.success(i18n.t("mcp:toast.reconnected"));
      } catch (err) {
        toast.error(i18n.t("mcp:toast.failedReconnect"), userFriendlyError(err));
        throw err;
      }
    },
    [api],
  );

  // ── Tools ──
  const listServerTools = useCallback(
    async (serverId: string): Promise<MCPToolInfo[]> => {
      const res = await api.listMCPServerTools({}, serverId);
      return res.tools ?? [];
    },
    [api],
  );

  // ── Grants ──
  const listServerGrants = useCallback(
    async (serverId: string) => {
      const res = await api.listMCPServerGrants({}, serverId);
      return res;
    },
    [api],
  );

  const grantAgent = useCallback(
    async (serverId: string, agentId: string, toolAllow?: string, toolDeny?: string) => {
      try {
        const res = await api.grantMCPServerAgent({}, { agent_id: agentId, tool_allow: toolAllow, tool_deny: toolDeny }, serverId);
        await invalidate();
        return res.grant;
      } catch (err) {
        toast.error(i18n.t("mcp:grants.failedGrant"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const revokeAgent = useCallback(
    async (serverId: string, agentId: string) => {
      try {
        await api.revokeMCPServerAgentGrant({}, serverId, agentId);
        await invalidate();
      } catch (err) {
        toast.error(i18n.t("mcp:grants.failedRevoke"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const listGrantsByAgent = useCallback(
    async (agentId: string): Promise<MCPAgentGrant[]> => {
      const res = await api.listAgentMCPServers({}, agentId);
      return res.grants ?? [];
    },
    [api],
  );

  // ── Refresh ref for external use ──
  if (!refreshRef.current) {
    refreshRef.current = async () => { await invalidate(); };
  }

  return {
    servers,
    loading,
    error,
    refresh: invalidate,
    createServer,
    updateServer,
    deleteServer,
    testConnection,
    reconnectServer,
    listServerTools,
    listServerGrants,
    grantAgent,
    revokeAgent,
    listGrantsByAgent,
  };
}