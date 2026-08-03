import { useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useHttp } from "@/hooks/use-ws";
import { queryKeys } from "@/lib/query-keys";
import { toast } from "@/stores/use-toast-store";
import i18next from "i18next";
import { userFriendlyError } from "@/lib/error-utils";
import type { SkillInfo } from "@/types/skill";

export type { SkillInfo };

export type SkillUploadResponse = {
  id?: string;
  slug: string;
  version: number;
  name: string;
  status?: string;
  is_new?: boolean;
  deps_warning?: string;
  deps_errors?: string[];
  missing_deps?: string[];
  deps_installed?: boolean;
};

export interface SkillAgentGrant {
  id: string;
  skill_id: number;
  agent_id: string;
  agent_name?: string;
  enabled: boolean;
  granted_by?: string;
  created_at: number;
}

export function useSkills() {
  const http = useHttp();
  const queryClient = useQueryClient();

  const { data: skills = [], isFetching: loading } = useQuery({
    queryKey: queryKeys.skills.all,
    queryFn: async () => {
      const res = await http.get<{ skills: SkillInfo[] }>("/v1/skills");
      return res.skills ?? [];
    },
    staleTime: 60_000,
  });

  const invalidate = useCallback(
    () => queryClient.invalidateQueries({ queryKey: queryKeys.skills.all }),
    [queryClient],
  );

  const getSkill = useCallback(
    async (id: string) => {
      const res = await http.get<{ skill: SkillInfo; content: string }>(`/v1/skills/${id}`);
      return { skill: res.skill, content: res.content ?? "" };
    },
    [http],
  );

  const uploadSkill = useCallback(
    async (file: File) => {
      const formData = new FormData();
      formData.append("file", file);
      const res = await http.upload<SkillUploadResponse>(
        "/v1/skills/upload",
        formData,
      );
      await invalidate();
      return res;
    },
    [http, invalidate],
  );

  const updateSkill = useCallback(
    async (id: string, updates: Record<string, unknown>) => {
      try {
        const res = await http.put<{ ok: string }>(`/v1/skills/${id}`, updates);
        await invalidate();
        toast.success(i18next.t("skills:toast.updated"));
        return res;
      } catch (err) {
        toast.error(i18next.t("skills:toast.updateFailed"), userFriendlyError(err));
        throw err;
      }
    },
    [http, invalidate],
  );

  const deleteSkill = useCallback(
    (id: string) => updateSkill(id, { status: "deleted" }),
    [updateSkill],
  );

  const listSkillGrants = useCallback(
    async (id: string) => {
      const res = await http.get<{ agent_grants: SkillAgentGrant[] }>(`/v1/skills/${id}/grants`);
      return res;
    },
    [http],
  );

  const grantAgent = useCallback(
    async (skillId: string, agentId: string) => {
      const res = await http.post<{ grant: SkillAgentGrant }>(`/v1/skills/${skillId}/grants/agent`, { agent_id: agentId });
      return res.grant;
    },
    [http],
  );

  const revokeAgent = useCallback(
    async (skillId: string, agentId: string) => {
      await http.delete(`/v1/skills/${skillId}/grants/agent/${agentId}`);
    },
    [http],
  );

  return {
    skills, loading, refresh: invalidate, getSkill,
    uploadSkill, updateSkill, deleteSkill,
    listSkillGrants, grantAgent, revokeAgent,
  };
}
