import { useCallback } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useHttp } from "@/hooks/use-ws";
import { queryKeys } from "@/lib/query-keys";
import type { SkillWithGrant } from "@/types/skill";
import { toast } from "@/stores/use-toast-store";
import i18next from "i18next";
import { userFriendlyError } from "@/lib/error-utils";

export function useAgentSkills(agentId: string) {
  const http = useHttp();
  const queryClient = useQueryClient();
  const queryKey = queryKeys.skills.agentGrants(agentId);

  const { data: skills = [], isLoading: loading } = useQuery({
    queryKey,
    queryFn: () =>
      http
        .get<{ skills: SkillWithGrant[] }>(`/v1/agents/${agentId}/skills`)
        .then((r) => r.skills ?? []),
    staleTime: 60_000,
  });

  const toggleSkillGrant = useCallback(
    async (skillId: string, enabled: boolean) => {
      queryClient.setQueryData<SkillWithGrant[]>(queryKey, (old) =>
        old?.map((s) => (s.id === skillId ? { ...s, granted: enabled } : s)),
      );
      try {
        await http.put(`/v1/skills/${skillId}/grant/${agentId}`, { enabled });
        toast.success(
          enabled
            ? i18next.t("agents:toast.skillGranted")
            : i18next.t("agents:toast.skillRevoked"),
        );
      } catch (err) {
        toast.error(
          i18next.t("agents:toast.skillGrantFailed"),
          userFriendlyError(err),
        );
        throw err;
      } finally {
        await queryClient.invalidateQueries({ queryKey });
      }
    },
    [http, agentId, queryClient, queryKey],
  );

  const grantSkill = useCallback(
    (skillId: string) => toggleSkillGrant(skillId, true),
    [toggleSkillGrant],
  );

  const revokeSkill = useCallback(
    (skillId: string) => toggleSkillGrant(skillId, false),
    [toggleSkillGrant],
  );

  return { skills, loading, grantSkill, revokeSkill, toggleSkillGrant };
}
