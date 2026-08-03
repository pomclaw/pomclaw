import { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { Trash2, Pencil, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ToolMultiSelect } from "./mcp-tool-multi-select";
import { useAgents } from "@/pages/agents/hooks/use-agents";
import type { MCPAgentGrant, MCPToolInfo } from "@/client/pomclawComponents";

interface MCPGrantsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  serverId: string;
  serverName: string;
  onLoadGrants: (serverId: string) => Promise<{
    agent_grants: MCPAgentGrant[];
  }>;
  onGrantAgent: (serverId: string, agentId: string, toolAllow?: string, toolDeny?: string) => Promise<MCPAgentGrant>;
  onRevokeAgent: (serverId: string, agentId: string) => Promise<void>;
  onLoadTools: (serverId: string) => Promise<MCPToolInfo[]>;
}

export function MCPGrantsDialog({
  open,
  onOpenChange,
  serverId,
  serverName,
  onLoadGrants,
  onGrantAgent,
  onRevokeAgent,
  onLoadTools,
}: MCPGrantsDialogProps) {
  const { t } = useTranslation("mcp");
  const { agents } = useAgents();
  const [grants, setGrants] = useState<MCPAgentGrant[]>([]);
  const [loading, setLoading] = useState(false);
  const [tools, setTools] = useState<MCPToolInfo[]>([]);
  const [editingGrant, setEditingGrant] = useState<MCPAgentGrant | null>(null);
  const [selectedAgentId, setSelectedAgentId] = useState("");
  const [toolAllow, setToolAllow] = useState<string[]>([]);
  const [toolDeny, setToolDeny] = useState<string[]>([]);
  const [saving, setSaving] = useState(false);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const [grantsRes, toolsRes] = await Promise.all([
        onLoadGrants(serverId),
        onLoadTools(serverId),
      ]);
      setGrants(grantsRes.agent_grants ?? []);
      setTools(toolsRes);
    } finally {
      setLoading(false);
    }
  }, [serverId, onLoadGrants, onLoadTools]);

  useEffect(() => {
    if (open) {
      loadData();
      resetForm();
    }
  }, [open, loadData]);

  const resetForm = () => {
    setEditingGrant(null);
    setSelectedAgentId("");
    setToolAllow([]);
    setToolDeny([]);
  };

  const startEdit = (grant: MCPAgentGrant) => {
    setEditingGrant(grant);
    setSelectedAgentId(grant.agent_id);
    setToolAllow(parseStringArray(grant.tool_allow));
    setToolDeny(parseStringArray(grant.tool_deny));
  };

  const handleSave = async () => {
    if (!selectedAgentId) return;
    setSaving(true);
    try {
      const allowStr = toolAllow.length > 0 ? JSON.stringify(toolAllow) : undefined;
      const denyStr = toolDeny.length > 0 ? JSON.stringify(toolDeny) : undefined;
      await onGrantAgent(serverId, selectedAgentId, allowStr, denyStr);
      await loadData();
      resetForm();
    } finally {
      setSaving(false);
    }
  };

  const handleRevoke = async (agentId: string) => {
    await onRevokeAgent(serverId, agentId);
    await loadData();
  };

  const toolOptions = tools.map((t) => ({ name: t.name, description: t.description }));

  const agentOptions = agents.filter((a) => !grants.some((g) => g.agent_id === a.id));

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("grants.title", { name: serverName })}</DialogTitle>
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
                        <div className="mt-1 flex flex-wrap gap-1">
                          {grant.tool_allow ? (
                            parseStringArray(grant.tool_allow).map((t) => (
                              <Badge key={t} variant="outline" className="text-xs font-mono">
                                {t}
                              </Badge>
                            ))
                          ) : (
                            <span className="text-xs text-muted-foreground">
                              {t("grants.allToolsAllowed")}
                            </span>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-1 ml-2">
                        <Button
                          variant="ghost"
                          size="icon-sm"
                          onClick={() => startEdit(grant)}
                        >
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
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

            {/* Add / Edit Form */}
            <div className="space-y-3 rounded-md border p-3">
              <h4 className="text-sm font-medium">
                {editingGrant ? t("grants.editGrant") : t("grants.addGrant")}
              </h4>

              <div className="space-y-2">
                <label className="text-sm text-muted-foreground">{t("grants.selectAgent")}</label>
                <Select
                  value={selectedAgentId}
                  onValueChange={setSelectedAgentId}
                  disabled={!!editingGrant}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t("grants.selectAgent")} />
                  </SelectTrigger>
                  <SelectContent>
                    {(editingGrant
                      ? agents.filter((a) => a.id === editingGrant.agent_id)
                      : agentOptions
                    ).map((agent) => (
                      <SelectItem key={agent.id} value={agent.id}>
                        {agent.display_name || agent.agent_key || agent.id}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <label className="text-sm text-muted-foreground">{t("grants.toolAllowList")}</label>
                <ToolMultiSelect
                  options={toolOptions}
                  value={toolAllow}
                  onChange={setToolAllow}
                  placeholder={t("grants.allowPlaceholder")}
                />
              </div>

              <div className="space-y-2">
                <label className="text-sm text-muted-foreground">{t("grants.toolDenyList")}</label>
                <ToolMultiSelect
                  options={toolOptions}
                  value={toolDeny}
                  onChange={setToolDeny}
                  placeholder={t("grants.denyPlaceholder")}
                />
              </div>

              <div className="flex justify-end gap-2">
                <Button variant="outline" size="sm" onClick={resetForm}>
                  {t("grants.cancel")}
                </Button>
                <Button size="sm" onClick={handleSave} disabled={!selectedAgentId || saving}>
                  {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
                  {editingGrant ? t("grants.update") : t("grants.grant")}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

function parseStringArray(val: string | undefined | null): string[] {
  if (!val) return [];
  try {
    const parsed = JSON.parse(val);
    if (Array.isArray(parsed)) return parsed;
    return [];
  } catch {
    return [];
  }
}