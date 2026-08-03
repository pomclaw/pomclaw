import { useCallback, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useApiClient } from "@/hooks/use-api-client";
import { queryKeys } from "@/lib/query-keys";
import { toast } from "@/stores/use-toast-store";
import i18n from "@/i18n";
import type {
  MemoryDocument,
  MemoryDocumentDetail,
  MemoryChunk,
  MemorySearchResult,
} from "@/types/memory";

export interface MemoryDocFilters {
  agentId?: string;
  userId?: string;
}

export function useMemoryDocuments(filters: MemoryDocFilters) {
  const api = useApiClient();
  const queryClient = useQueryClient();

  const queryKey = queryKeys.memory.list({ ...filters });

  const { data, isLoading, isFetching } = useQuery({
    queryKey,
    queryFn: async () => {
      if (!filters.agentId) {
        const res = await api.listMemoryDocuments();
        return (res as { documents?: MemoryDocument[] }).documents ?? [];
      }
      const params = filters.userId
        ? ({ user_id: filters.userId } as Parameters<typeof api.getAgentMemoryDocuments>[0])
        : {};
      const res = await api.getAgentMemoryDocuments(params, filters.agentId);
      return (res as { documents?: MemoryDocument[] }).documents ?? [];
    },
    placeholderData: (prev) => prev,
    staleTime: 60_000,
  });

  const documents = data ?? [];

  const invalidate = useCallback(
    () => queryClient.invalidateQueries({ queryKey: queryKeys.memory.all }),
    [queryClient],
  );

  const getDocument = useCallback(
    async (documentID: number) => {
      const res = await api.getMemoryDocument({}, filters.agentId!, documentID);
      return (res as { document?: MemoryDocumentDetail }).document;
    },
    [api, filters.agentId],
  );

  const updateDocument = useCallback(
    async (documentID: number, content: string) => {
      try {
        await api.putMemoryDocument({}, { content }, filters.agentId!, documentID);
        await invalidate();
        toast.success(i18n.t("memory:toast.docUpdated"));
      } catch (err) {
        toast.error(i18n.t("memory:toast.docUpdateFailed"), err instanceof Error ? err.message : i18n.t("memory:toast.unknownError"));
        throw err;
      }
    },
    [api, filters.agentId, invalidate],
  );

  const deleteDocument = useCallback(
    async (documentID: number) => {
      if (!filters.agentId) {
        toast.error(i18n.t("memory:toast.docDeleteFailed"), "No agent selected");
        return;
      }
      try {
        await api.deleteMemoryDocument({}, filters.agentId, documentID);
        await invalidate();
        toast.success(i18n.t("memory:toast.docDeleted"));
      } catch (err) {
        toast.error(i18n.t("memory:toast.docDeleteFailed"), err instanceof Error ? err.message : i18n.t("memory:toast.unknownError"));
        throw err;
      }
    },
    [api, filters.agentId, invalidate],
  );

  const getChunks = useCallback(
    async (path: string, userId?: string) => {
      const params = { path, ...(userId ? { user_id: userId } : {}) } as Parameters<typeof api.listMemoryChunks>[0];
      const res = await api.listMemoryChunks(params, filters.agentId!);
      return (res as { chunks?: MemoryChunk[] }).chunks ?? [];
    },
    [api, filters.agentId],
  );

  const createDocument = useCallback(
    async (path: string, content: string) => {
      try {
        await api.putMemoryDocument({}, { path, content }, filters.agentId!, 0);
        await invalidate();
        toast.success(i18n.t("memory:toast.docCreated", { defaultValue: "文档已创建" }));
      } catch (err) {
        toast.error(i18n.t("memory:toast.failedCreate"), err instanceof Error ? err.message : i18n.t("memory:toast.unknownError"));
        throw err;
      }
    },
    [api, filters.agentId, invalidate],
  );

  const indexDocument = useCallback(
    async (path: string) => {
      try {
        await api.indexDocument({}, { path }, filters.agentId!);
        toast.success(i18n.t("memory:toast.docIndexed"), path);
      } catch (err) {
        toast.error(i18n.t("memory:toast.docIndexFailed"), err instanceof Error ? err.message : i18n.t("memory:toast.unknownError"));
        throw err;
      }
    },
    [api, filters.agentId],
  );

  const indexAll = useCallback(
    async () => {
      try {
        await api.indexAll({}, filters.agentId!);
        toast.success(i18n.t("memory:toast.allIndexed"));
      } catch (err) {
        toast.error(i18n.t("memory:toast.allIndexFailed"), err instanceof Error ? err.message : i18n.t("memory:toast.unknownError"));
        throw err;
      }
    },
    [api, filters.agentId],
  );

  return {
    documents,
    loading: isLoading,
    fetching: isFetching,
    refresh: invalidate,
    getDocument,
    createDocument,
    updateDocument,
    deleteDocument,
    getChunks,
    indexDocument,
    indexAll,
  };
}

export function useMemorySearch(agentId: string) {
  const api = useApiClient();
  const [results, setResults] = useState<MemorySearchResult[]>([]);
  const [searching, setSearching] = useState(false);

  const search = useCallback(
    async (query: string, _userId?: string, maxResults?: number) => {
      setSearching(true);
      try {
        const res = await api.searchMemory({}, { query, limit: maxResults || 10 }, agentId);
        const items = (res as unknown as { results?: MemorySearchResult[] }).results ?? [];
        setResults(items);
        return items;
      } catch (err) {
        toast.error(i18n.t("memory:toast.searchFailed"), err instanceof Error ? err.message : i18n.t("memory:toast.unknownError"));
        setResults([]);
        return [];
      } finally {
        setSearching(false);
      }
    },
    [api, agentId],
  );

  return { results, searching, search };
}
