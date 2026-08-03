import { useState, useEffect, useCallback } from "react";
import { useApiClient } from "@/hooks/use-api-client";
import type { SessionInfo } from "@/types/session";
import type { Session } from "@/client/pomclawComponents";
import { toast } from "@/stores/use-toast-store";
import i18next from "i18next";
import { userFriendlyError } from "@/lib/error-utils";
import { uniqueId } from "@/lib/utils";

/**
 * Convert HTTP API Session to SessionInfo format.
 */
function sessionToSessionInfo(session: Session): SessionInfo {
  return {
    key: `${session.id}`,
    agentId: session.agent_id,
    messageCount: session.message_count,
    created: session.created,
    updated: session.updated,
    label: session.title || '新对话',
  };
}

/**
 * Manages the session list for the chat sidebar.
 * Loads all sessions (across all agents) using HTTP API.
 */
export function useChatSessions() {
  const api = useApiClient();
  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadSessions = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      // 加载全部 session，不过滤 agent（API 已移除 agent_id 参数）
      const res = await api.listSessions({ offset: 0, limit: 50 });
      const converted = res.sessions.map((s) => sessionToSessionInfo(s));
      // API 已返回按 updated_at 倒序的结果，但保险起见再排序一次
      const sorted = converted.sort(
        (a: SessionInfo, b: SessionInfo) =>
          new Date(b.updated).getTime() - new Date(a.updated).getTime(),
      );
      setSessions(sorted);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load sessions");
    } finally {
      setLoading(false);
    }
  }, [api]);

  useEffect(() => {
    loadSessions();
  }, [loadSessions, api]);

  const buildNewSessionKey = useCallback(() => {
    // Return a special marker for new sessions that haven't been saved yet
    return `new:${uniqueId()}`;
  }, []);

  const deleteSession = useCallback(async (key: string) => {
    try {
      await api.deleteSession({}, Number(key));
      await loadSessions();
      toast.success(i18next.t("sessions:toast.deleted"));
    } catch (err) {
      toast.error(i18next.t("sessions:toast.deleteFailed"), userFriendlyError(err));
      throw err;
    }
  }, [api, loadSessions]);

  return {
    sessions,
    loading,
    error,
    refresh: loadSessions,
    buildNewSessionKey,
    deleteSession,
  };
}
