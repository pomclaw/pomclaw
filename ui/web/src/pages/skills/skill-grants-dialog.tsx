import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Trash2, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAgents } from "@/pages/agents/hooks/use-agents";
import type { SkillAgentGrant } from "./hooks/use-skills";

interface SkillGrantsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  skillId: string;
  skillName: string;
  onLoadGrants: (skillId: string) => Promise<{
    agent_grants: SkillAgentGrant[];
  }>;
  onGrantAgent: (skillId: string, agentId: string) => Promise<SkillAgentGrant>;
  onRevokeAgent: (skillId: string, agentId: string) => Promise<void>;
}

export function SkillGrantsDialog({
  open,
  onOpenChange,
  skillId,
  skillName,
  onLoadGrants,
  onGrantAgent,
  onRevokeAgent,
}: SkillGrantsDialogProps) {
  const { t } = useTranslation("skills");
  const { agents } = useAgents();
  const [grants, setGrants] = useState<SkillAgentGrant[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState("");
  const [saving, setSaving] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await onLoadGrants(skillId);
      setGrants(res.agent_grants ?? []);
    } finally {
      setLoading(false);
    }
  }, [skillId, onLoadGrants]);

  useEffect(() => {
    if (open) {
      loadData();
      setSelectedAgentId("");
    }
  }, [open, loadData]);

  const handleGrant = async () => {
    if (!selectedAgentId) return;
    setSaving(true);
    try {
      await onGrantAgent(skillId, selectedAgentId);
      await loadData();
      setSelectedAgentId("");
    } finally {
      setSaving(false);
    }
  };

  const handleRevoke = async (agentId: string) => {
    await onRevokeAgent(skillId, agentId);
    await loadData();
  };

  const agentOptions = agents.filter((a) => !grants.some((g) => g.agent_id === a.id));

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("grants.title", { name: skillName })}</DialogTitle>
        </DialogHeader>

        {loading && (
          <div className="flex justify-center py-4">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </div>
        )}

        {!loading && (
          <>
            {/* Existing Grants */}
            {grants.length > 0 && (
              <div className="space-y-2">
                <h4 className="text-sm font-medium">{t("grants.currentGrants")}</h4>
                {grants.map((grant) => {
                  const agent = agents.find((a) => a.id === grant.agent_id);
                  return (
                    <div
                      key={grant.id}
                      className="flex items-center justify-between rounded-md border p-2.5"
                    >
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium truncate">
                          {agent?.display_name || agent?.agent_key || grant.agent_id}
                        </p>
                      </div>
                      <div className="flex items-center gap-1 ml-2">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => handleRevoke(grant.agent_id)}
                        >
                          <Trash2 className="h-3.5 w-3.5 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Add Grant Form */}
            <div className="space-y-3 rounded-md border p-3">
              <h4 className="text-sm font-medium">{t("grants.addGrant")}</h4>

              <div className="space-y-2">
                <label className="text-sm text-muted-foreground">{t("grants.selectAgent")}</label>
                <Select
                  value={selectedAgentId}
                  onValueChange={setSelectedAgentId}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t("grants.selectAgent")} />
                  </SelectTrigger>
                  <SelectContent>
                    {agentOptions.map((agent) => (
                      <SelectItem key={agent.id} value={agent.id}>
                        {agent.display_name || agent.agent_key || agent.id}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={() => setSelectedAgentId("")}>
                  {t("grants.cancel")}
                </Button>
                <Button size="sm" onClick={handleGrant} disabled={!selectedAgentId || saving}>
                  {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                  {t("grants.grant")}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}