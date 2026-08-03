import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import type {
  AgentData, MemoryConfig,
} from "@/types/agent";
import { StickySaveBar } from "@/components/shared/sticky-save-bar";
import { ModelBudgetSection } from "./overview-sections/model-budget-section";
import { EvolutionSection } from "./overview-sections/evolution-section";
import { PromptSettingsSection } from "./overview-sections/prompt-settings-section";
import { ChatGPTOAuthRoutingSummarySection } from "./overview-sections/chatgpt-oauth-routing-summary-section";
import { ShareSection } from "./overview-sections/share-section";
import { HeartbeatCard } from "./overview-sections/heartbeat-card";
import { MemorySection } from "./config-sections";
import type { UseAgentHeartbeatReturn } from "../hooks/use-agent-heartbeat";

interface AgentOverviewTabProps {
  agent: AgentData;
  onUpdate: (updates: Record<string, unknown>) => Promise<void>;
  heartbeat: UseAgentHeartbeatReturn;
  onManageCodexPool: () => void;
}

export function AgentOverviewTab({ agent, onUpdate, heartbeat, onManageCodexPool }: AgentOverviewTabProps) {
  const { t } = useTranslation("agents");

  // Model & Budget
  const [provider, setProvider] = useState(agent.provider || "");
  const [model, setModel] = useState(agent.model || "");
  const [contextWindow, setContextWindow] = useState(agent.context_window || 200000);
  const [maxToolIterations, setMaxToolIterations] = useState(agent.max_tool_iterations || 20);
  const [budgetDollars, setBudgetDollars] = useState(
    agent.budget_monthly_cents ? String(agent.budget_monthly_cents / 100) : "",
  );
  // Evolution (predefined only)
  const [selfEvolve, setSelfEvolve] = useState(Boolean(agent.self_evolve));
  const [skillEvolve, setSkillEvolve] = useState(Boolean(agent.skill_evolve));
  const [skillNudgeInterval, setSkillNudgeInterval] = useState(
    typeof agent.skill_nudge_interval === "number" ? agent.skill_nudge_interval : 15,
  );
  // Sharing
  const [isShared, setIsShared] = useState(Boolean(agent.is_shared));

  // Memory (always shown — per-agent overrides, empty = use system defaults)
  const [mem, setMem] = useState<MemoryConfig>(agent.memory_config ?? {});

  // Save state
  const [saving, setSaving] = useState(false);
  const [llmSaveBlocked, setLlmSaveBlocked] = useState(false);

  // Sync all state when agent data changes (e.g., after refresh or initial load)
  useEffect(() => {
    setProvider(agent.provider || "");
    setModel(agent.model || "");
    setContextWindow(agent.context_window || 200000);
    setMaxToolIterations(agent.max_tool_iterations || 20);
    setBudgetDollars(
      agent.budget_monthly_cents ? String(agent.budget_monthly_cents / 100) : "",
    );
    setSelfEvolve(Boolean(agent.self_evolve));
    setSkillEvolve(Boolean(agent.skill_evolve));
    setSkillNudgeInterval(
      typeof agent.skill_nudge_interval === "number" ? agent.skill_nudge_interval : 15,
    );
    setIsShared(Boolean(agent.is_shared));
    setMem(agent.memory_config ?? {});
  }, [agent]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const budgetCents = budgetDollars ? Math.round(parseFloat(budgetDollars) * 100) : null;
      const updates: Record<string, unknown> = {
        provider,
        model,
        context_window: contextWindow,
        max_tool_iterations: maxToolIterations,
        budget_monthly_cents: budgetCents,
        memory_config: mem,
        // Promoted fields sent at top level (NOT NULL columns — send "" not null)
        self_evolve: selfEvolve,
        skill_evolve: skillEvolve,
        skill_nudge_interval: skillEvolve ? skillNudgeInterval : 15,
        is_shared: isShared,
      };
      // When the provider changes, clear stale pool routing config so it
      // doesn't reference members from the previous provider's pool.
      if (provider !== agent.provider) {
        updates.chatgpt_oauth_routing = null;
      }
      await onUpdate(updates);
    } catch {
      // toast shown by hook
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-4">
      <PromptSettingsSection agent={agent} onUpdate={onUpdate} />

      <ModelBudgetSection
        provider={provider}
        onProviderChange={setProvider}
        model={model}
        onModelChange={setModel}
        contextWindow={contextWindow}
        onContextWindowChange={setContextWindow}
        maxToolIterations={maxToolIterations}
        onMaxToolIterationsChange={setMaxToolIterations}
        savedProvider={agent.provider}
        savedModel={agent.model}
        budgetDollars={budgetDollars}
        onBudgetDollarsChange={setBudgetDollars}
        onSaveBlockedChange={setLlmSaveBlocked}
      />

      <ShareSection
        isShared={isShared}
        onIsSharedChange={setIsShared}
      />

      <ChatGPTOAuthRoutingSummarySection agent={agent} onManage={onManageCodexPool} />
      {provider !== agent.provider && !!agent.chatgpt_oauth_routing && (
        <p className="text-xs text-amber-600 dark:text-amber-400 -mt-2 px-1">
          {t("chatgptOAuthRouting.providerChangedWarning")}
        </p>
      )}

      {agent.agent_type === "predefined" && (
        <EvolutionSection
          agentId={agent.id}
          selfEvolve={selfEvolve}
          onSelfEvolveChange={setSelfEvolve}
          skillEvolve={skillEvolve}
          onSkillEvolveChange={setSkillEvolve}
          skillNudgeInterval={skillNudgeInterval}
          onSkillNudgeIntervalChange={setSkillNudgeInterval}
        />
      )}

      {/* Memory — always visible, per-agent overrides */}
      <MemorySection
        value={mem}
        onChange={setMem}
      />

      <HeartbeatCard heartbeat={heartbeat} />

      <StickySaveBar
        onSave={handleSave}
        saving={saving}
        disabled={llmSaveBlocked}
        label={t("general.saveChanges")}
        savingLabel={t("general.saving")}
      />
    </div>
  );
}