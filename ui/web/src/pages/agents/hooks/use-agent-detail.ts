import { useEffect, useCallback, useState } from "react";
import { useApiClient } from "@/hooks/use-api-client";
import { toast } from "@/stores/use-toast-store";
import i18n from "@/i18n";
import { userFriendlyError } from "@/lib/error-utils";
import type { AgentData, BootstrapFile } from "@/types/agent";

export function useAgentDetail(agentId: string | undefined) {
  const api = useApiClient();

  const [agent, setAgent] = useState<AgentData | null>(null);
  const [files, setFiles] = useState<BootstrapFile[]>([]);
  const [loading, setLoading] = useState(false);

  // Fetch agent and files every time agentId changes (no caching)
  useEffect(() => {
    if (!agentId) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        // Fetch agent from API
        const resp = await api.getAgent({}, agentId);
        const agentData = resp.agent;
        setAgent(agentData);

        // Load files via HTTP API
        let filesData: BootstrapFile[] = [];
        try {
          const filesRes = await api.listAgentFiles({}, agentId);
          filesData = filesRes.files ?? [];
        } catch {
          // ignore
        }
        setFiles(filesData);
      } catch {
        // On error, set minimal agent data
        setAgent({
          id: agentId,
          agent_key: agentId,
          owner_id: "",
          provider: "",
          model: "",
          context_window: 0,
          max_tool_iterations: 0,
          workspace: "",
          restrict_to_workspace: false,
          agent_type: "open" as const,
          is_default: false,
          status: "active",
          is_shared: false,
          self_evolve: false,
          skill_evolve: false,
          emoji: null,
          agent_description: null,
          thinking_level: null,
          max_tokens: null,
          skill_nudge_interval: null,
          tools_config: null,
          sandbox_config: null,
          subagents_config: null,
          memory_config: null,
          compaction_config: null,
          context_pruning: null,
          other_config: null,
          budget_monthly_cents: null,
        });
        setFiles([]);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [agentId, api]);

  const updateAgent = useCallback(
    async (updates: Record<string, unknown>) => {
      if (!agentId) return;
      try {
        await api.updateAgent({}, updates as any, agentId);
        toast.success(i18n.t("agents:toast.updated"));
        // Refetch data after update
        const resp = await api.getAgent({}, agentId);
        setAgent(resp.agent);
      } catch (err) {
        toast.error(i18n.t("agents:toast.updateFailed"), userFriendlyError(err));
        throw err;
      }
    },
    [agentId, api],
  );

  const getFile = useCallback(
    async (name: string): Promise<BootstrapFile | null> => {
      if (!agent) return null;
      const res = await api.getAgentFile({}, agentId, name);
      return res.file ?? null;
    },
    [agent, agentId, api],
  );

  const setFile = useCallback(
    async (name: string, content: string) => {
      if (!agent) return;
      try {
        await api.setAgentFile({}, { content }, agentId, name);
        toast.success(i18n.t("agents:toast.updated"));
        // Refetch files after set
        const filesRes = await api.listAgentFiles({}, agentId);
        setFiles(filesRes.files ?? []);
      } catch (err) {
        toast.error(i18n.t("agents:toast.updateFailed"), userFriendlyError(err));
        throw err;
      }
    },
    [agent, agentId, api],
  );

  const regenerateAgent = useCallback(
    async (_prompt: string) => {
      if (!agentId) return;
      // Note: regenerate API method not yet generated
    },
    [agentId],
  );

  const resummonAgent = useCallback(async () => {
    if (!agentId) return;
    // Note: resummon API method not yet generated
  }, [agentId]);

  const deleteAgent = useCallback(async () => {
    if (!agentId) return;
    await api.deleteAgent({}, agentId);
  }, [agentId, api]);

  const refresh = useCallback(async () => {
    if (!agentId) return;
    try {
      const resp = await api.getAgent({}, agentId);
      setAgent(resp.agent);
    } catch {
      // ignore
    }
  }, [agentId, api]);

  return { agent, files, loading, updateAgent, getFile, setFile, regenerateAgent, resummonAgent, deleteAgent, refresh };
}
