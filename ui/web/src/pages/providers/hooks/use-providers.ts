import { useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import i18next from "i18next";
import { useApiClient } from "@/hooks/use-api-client";
import type { Provider, CreateProviderReq, UpdateProviderReq } from "@/client/pomclawComponents";
import { queryKeys } from "@/lib/query-keys";
import { toast } from "@/stores/use-toast-store";

export type { Provider };

export function useProviders(enabled = true) {
  const api = useApiClient();
  const queryClient = useQueryClient();

  const { data: providers = [], isLoading: loading } = useQuery({
    queryKey: queryKeys.providers.all,
    enabled,
    queryFn: async () => {
      const res = await api.listProviders();
      return res.providers ?? [];
    },
    staleTime: 60_000,
  });

  const invalidate = useCallback(
    () => queryClient.invalidateQueries({ queryKey: queryKeys.providers.all }),
    [queryClient],
  );

  const createProvider = useCallback(
    async (data: CreateProviderReq) => {
      try {
        const res = await api.createProvider(data);
        await invalidate();
        toast.success(
          i18next.t("providers:toast.created"),
          i18next.t("providers:toast.createdDesc", { name: data.name }),
        );
        return res.provider;
      } catch (err) {
        toast.error(i18next.t("providers:toast.failedCreate"), err instanceof Error ? err.message : "");
        throw err;
      }
    },
    [api, invalidate],
  );

  const updateProvider = useCallback(
    async (id: number, data: Partial<UpdateProviderReq>) => {
      try {
        await api.updateProvider({}, data as UpdateProviderReq, id);
        await invalidate();
        toast.success(i18next.t("providers:toast.updated"));
      } catch (err) {
        toast.error(i18next.t("providers:toast.failedUpdate"), err instanceof Error ? err.message : "");
        throw err;
      }
    },
    [api, invalidate],
  );

  const deleteProvider = useCallback(
    async (id: number) => {
      try {
        await api.deleteProvider({}, id);
        await invalidate();
        toast.success(i18next.t("providers:toast.deleted"));
      } catch (err) {
        toast.error(i18next.t("providers:toast.failedDelete"), err instanceof Error ? err.message : "");
        throw err;
      }
    },
    [api, invalidate],
  );

  return {
    providers,
    loading,
    refresh: invalidate,
    createProvider,
    updateProvider,
    deleteProvider,
  };
}
