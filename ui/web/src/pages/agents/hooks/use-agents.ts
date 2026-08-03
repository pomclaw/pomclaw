import { useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useWs, useHttp } from "@/hooks/use-ws";
import { useApiClient } from "@/hooks/use-api-client";
import { useAuthStore } from "@/stores/use-auth-store";
import { Methods } from "@/api/protocol";
import { queryKeys } from "@/lib/query-keys";
import { toast } from "@/stores/use-toast-store";
import i18n from "@/i18n";
import { userFriendlyError } from "@/lib/error-utils";
import type { AgentData } from "@/types/agent";
import type { Agent, CreateAgentReq, UpdateAgentReq, Provider } from "@/client/pomclawComponents";

function toAgentData(a: Agent): AgentData {
  return {
    ...a,
    agent_key: a.id,
    agent_type: a.agent_type as AgentData["agent_type"],
    provider: "",
    created_at: a.created_at ? String(a.created_at) : undefined,
    updated_at: a.updated_at ? String(a.updated_at) : undefined,
  };
}

export function useAgents() {
  const ws = useWs();
  const http = useHttp();
  const api = useApiClient();
  const connected = useAuthStore((s) => s.connected);
  const queryClient = useQueryClient();

  const { data: agents = [], isPending: loading, error: queryError } = useQuery({
    queryKey: queryKeys.agents.all,
    queryFn: async () => {
      try {
        const res = await api.listAgents();
        if (res.agents && res.agents.length > 0) {
          return res.agents.map(toAgentData);
        }
      } catch {
        // HTTP may fail if user doesn't have access — fall through to WS
      }

      if (!ws.isConnected) return [];
      const res = await ws.call<{ agents: Array<{ id: string; model: string; isRunning: boolean; agentType?: string }> }>(Methods.AGENTS_LIST);
      return (res.agents ?? []).map((a): AgentData => ({
        id: a.id,
        agent_key: a.id,
        owner_id: "",
        provider: "",
        model: a.model,
        context_window: 0,
        max_tool_iterations: 0,
        workspace: "",
        restrict_to_workspace: false,
        agent_type: a.agentType === "predefined" ? "predefined" : "open",
        is_default: false,
        status: a.isRunning ? "active" : "inactive",
      }));
    },
    staleTime: 60_000,
    enabled: connected,
  });

  const error = queryError instanceof Error ? queryError.message : queryError ? "Failed to load agents" : null;

  const invalidate = useCallback(
    () => queryClient.invalidateQueries({ queryKey: queryKeys.agents.all }),
    [queryClient],
  );

  const createAgent = useCallback(
    async (data: Partial<AgentData>) => {
      try {
        // Validate required fields
        if (!data.display_name || !data.provider_id || !data.model) {
          throw new Error("Missing required fields: display_name, provider_id, or model");
        }

        // Transform data to match CreateAgentReq type
        const req: CreateAgentReq = {
          display_name: data.display_name,
          provider_id: data.provider_id,
          model: data.model,
          emoji: data.emoji ?? undefined,
          agent_description: data.agent_description ?? undefined,
          self_evolve: data.self_evolve ?? false,
          is_shared: data.is_shared ?? false,
          thinking_level: data.thinking_level ?? undefined,
          max_tokens: data.max_tokens ?? undefined,
          skill_evolve: data.skill_evolve ?? undefined,
          context_window: data.context_window ?? undefined,
          max_tool_iterations: data.max_tool_iterations ?? undefined,
          workspace: data.workspace ?? undefined,
        };
        const res = await api.createAgent(req);
        await invalidate();
        toast.success(i18n.t("agents:toast.created"), `${data.display_name || "Agent"} has been added`);
        return res;
      } catch (err) {
        toast.error(i18n.t("agents:toast.createFailed"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const updateAgent = useCallback(
    async (id: string, data: Partial<AgentData>) => {
      try {
        await api.updateAgent({}, data as UpdateAgentReq, id);
        await invalidate();
        toast.success(i18n.t("agents:toast.updated"), `${data.display_name || data.agent_key || "Agent"} has been updated`);
      } catch (err) {
        toast.error(i18n.t("agents:toast.updateFailed"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const deleteAgent = useCallback(
    async (id: string) => {
      try {
        await api.deleteAgent({}, id);
        await invalidate();
        toast.success(i18n.t("agents:toast.deleted"));
      } catch (err) {
        toast.error(i18n.t("agents:toast.deleteFailed"), userFriendlyError(err));
        throw err;
      }
    },
    [api, invalidate],
  );

  const resummonAgent = useCallback(
    async (id: string) => {
      await http.post(`/v1/agents/${id}/resummon`);
    },
    [http],
  );

  const cancelSummonAgent = useCallback(
    async (id: string) => {
      await http.post(`/v1/agents/${id}/cancel-summon`);
    },
    [http],
  );

  return { agents, loading, error, refresh: invalidate, createAgent, updateAgent, deleteAgent, resummonAgent, cancelSummonAgent };
}
