import { useState, useEffect, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Server, Zap, Wrench } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { useApiClient } from "@/hooks/use-api-client";
import { ToolPolicySection } from "./config-sections";
import { useBuiltinTools } from "@/pages/builtin-tools/hooks/use-builtin-tools";
import type { ToolPolicyConfig } from "@/types/agent";
import { Button } from "@/components/ui/button";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { Badge } from "@/components/ui/badge";
import type { AgentAuthorizationSkill, AgentAuthorizationMCP } from "@/client/pomclawComponents";

interface Props {
  agentId: string;
  agent: { tools_config?: ToolPolicyConfig | null };
  onUpdate: (updates: Record<string, unknown>) => Promise<void>;
}

export function AgentSkillsTab({ agentId, agent, onUpdate }: Props) {
  const { t } = useTranslation("agents");
  const api = useApiClient();

  // ── Consolidated authorizations query ──
  const { data: auth, isLoading: authLoading } = useQuery({
    queryKey: ["agent-authorizations", agentId],
    queryFn: () => api.getAgentAuthorizations({}, agentId),
  });

  // ── Built-in tools ──
  const { tools: builtinTools, loading: toolsLoading } = useBuiltinTools();

  // Group built-in tools by category
  const toolsByCategory = useMemo(() => {
    const map = new Map<string, typeof builtinTools>();
    for (const t of builtinTools) {
      const cat = t.category || "other";
      if (!map.has(cat)) map.set(cat, []);
      map.get(cat)!.push(t);
    }
    return Array.from(map.entries());
  }, [builtinTools]);

  // ── Tool Policy (editable, saved independently) ──
  const [toolsEnabled, setToolsEnabled] = useState(agent.tools_config != null);
  const [tools, setTools] = useState<ToolPolicyConfig>(agent.tools_config ?? {});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setToolsEnabled(agent.tools_config != null);
    setTools(agent.tools_config ?? {});
  }, [agent.tools_config]);

  const toolsDirty =
    toolsEnabled !== (agent.tools_config != null) ||
    JSON.stringify(tools) !== JSON.stringify(agent.tools_config ?? {});

  const handleSaveTools = async () => {
    setSaving(true);
    try {
      await onUpdate({
        tools_config: toolsEnabled
          ? { profile: tools.profile, allow: tools.allow, deny: tools.deny, alsoAllow: tools.alsoAllow, byProvider: tools.byProvider }
          : {},
      });
    } finally {
      setSaving(false);
    }
  };

  // ── Derived data from consolidated endpoint ──
  const mcpServers: AgentAuthorizationMCP[] = auth?.mcp_servers ?? [];
  const skills: AgentAuthorizationSkill[] = auth?.skills ?? [];

  return (
    <div className="space-y-6">
      {/* Tool Policy — editable */}
      <section className="rounded-lg border p-3 sm:p-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium">{t("configSections.toolPolicy.title")}</h3>
          {toolsDirty && (
            <Button size="sm" onClick={handleSaveTools} disabled={saving}>
              {saving ? t("general.saving") : t("general.saveChanges")}
            </Button>
          )}
        </div>
        <ToolPolicySection
          enabled={toolsEnabled}
          value={tools}
          onToggle={(v: boolean) => { setToolsEnabled(v); if (!v) setTools({}); }}
          onChange={setTools}
        />
      </section>

      {/* Built-in tool list — read-only */}
      <section className="space-y-3 rounded-lg border p-3 sm:p-4">
        <div className="flex items-center gap-2">
          <Wrench className="h-4 w-4 text-muted-foreground" />
          <h3 className="text-sm font-medium">{t("detail.skillsTab.builtinTools", "Built-in Tools")}</h3>
          {!toolsLoading && (
            <span className="text-xs text-muted-foreground">({builtinTools.length})</span>
          )}
        </div>
        {toolsLoading ? (
          <TableSkeleton />
        ) : toolsByCategory.length === 0 ? (
          <p className="text-xs text-muted-foreground italic px-1">
            {t("detail.skillsTab.noTools", "No tools available.")}
          </p>
        ) : (
          <div className="space-y-3">
            {toolsByCategory.map(([category, tools]) => (
              <div key={category}>
                <h4 className="text-2xs font-semibold text-muted-foreground uppercase tracking-wider mb-1 px-1">
                  {category}
                </h4>
                <div className="divide-y rounded-lg border max-h-[300px] overflow-y-auto overscroll-contain">
                  {tools.map((tool) => (
                    <div key={tool.name} className="flex items-center justify-between gap-3 px-3 py-2.5">
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-1.5">
                          <span className="text-sm font-medium truncate">{tool.display_name}</span>
                          <code className="text-2xs text-muted-foreground shrink-0">{tool.name}</code>
                        </div>
                        {tool.description && (
                          <p className="mt-0.5 truncate text-xs text-muted-foreground">{tool.description}</p>
                        )}
                      </div>
                      <Badge variant="secondary" className="text-2xs shrink-0">
                        {tool.enabled ? t("skills.granted") : t("skills.notGranted")}
                      </Badge>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* MCP list — read-only */}
      <section className="space-y-3 rounded-lg border p-3 sm:p-4">
        <div className="flex items-center gap-2">
          <Server className="h-4 w-4 text-muted-foreground" />
          <h3 className="text-sm font-medium">{t("detail.skillsTab.mcpServers", "MCP Servers")}</h3>
          {!authLoading && (
            <span className="text-xs text-muted-foreground">({mcpServers.length})</span>
          )}
        </div>
        {authLoading ? (
          <TableSkeleton />
        ) : mcpServers.length === 0 ? (
          <p className="text-xs text-muted-foreground italic px-1">
            {t("detail.skillsTab.noMCP", "No MCP servers granted to this agent.")}
          </p>
        ) : (
          <div className="divide-y rounded-lg border max-h-[300px] overflow-y-auto overscroll-contain">
            {mcpServers.map((server) => (
              <div key={server.server_id} className="flex items-center justify-between gap-3 px-3 py-2.5">
                <div className="min-w-0 flex-1">
                  <span className="text-sm font-medium">{server.server_name}</span>
                  {server.transport && (
                    <p className="mt-0.5 text-xs text-muted-foreground">{server.transport}</p>
                  )}
                </div>
                <Badge variant="secondary" className="text-2xs shrink-0">
                  {server.enabled ? t("skills.granted") : t("skills.notGranted")}
                </Badge>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Skills list — read-only */}
      <section className="space-y-3 rounded-lg border p-3 sm:p-4">
        <div className="flex items-center gap-2">
          <Zap className="h-4 w-4 text-amber-500" />
          <h3 className="text-sm font-medium">{t("detail.skills")}</h3>
          {!authLoading && (
            <span className="text-xs text-muted-foreground">({skills.length})</span>
          )}
        </div>
        {authLoading && skills.length === 0 ? (
          <TableSkeleton />
        ) : skills.length === 0 ? (
          <p className="text-xs text-muted-foreground italic px-1">{t("skills.noSkillsAvailable")}</p>
        ) : (
          <div className="divide-y rounded-lg border max-h-[300px] overflow-y-auto overscroll-contain">
            {skills.map((skill) => (
              <div key={skill.id} className="flex items-center justify-between gap-3 px-3 py-2.5">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-1.5">
                    <span className="text-sm font-medium truncate">{skill.name}</span>
                    <Badge variant="outline" className="text-2xs shrink-0">
                      {skill.visibility}
                    </Badge>
                    {skill.is_system && (
                      <Badge variant="outline" className="border-blue-500 text-blue-600 text-2xs shrink-0">
                        {t("skills.system")}
                      </Badge>
                    )}
                  </div>
                  {skill.description && (
                    <p className="mt-0.5 truncate text-xs text-muted-foreground">{skill.description}</p>
                  )}
                </div>
                <Badge variant="secondary" className="text-2xs shrink-0">
                  {t("skills.granted")}
                </Badge>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}