import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Bot } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useAgents } from "@/pages/agents/hooks/use-agents";
import type { AgentData } from "@/types/agent";

interface AgentPickerModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAgentSelected: (agentId: string) => void;
}

/** Extract emoji from agent */
function agentEmoji(agent: AgentData): string | undefined {
  return agent.emoji || undefined;
}

export function AgentPickerModal({
  open,
  onOpenChange,
  onAgentSelected,
}: AgentPickerModalProps) {
  const { t } = useTranslation("chat");
  const { agents: allAgents } = useAgents();

  // Filter to active agents
  const agents = useMemo(
    () => (allAgents ?? []).filter((a) => a.status === "active"),
    [allAgents],
  );

  const handleAgentSelect = (agentId: string) => {
    onAgentSelected(agentId);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{t("selectAgent.title")}</DialogTitle>
          <DialogDescription>
            {t("selectAgent.description")}
          </DialogDescription>
        </DialogHeader>

        <div className="max-h-[60vh] overflow-y-auto space-y-2">
          {agents.length === 0 && (
            <div className="px-4 py-6 text-center text-sm text-muted-foreground">
              {t("noAgentsAvailable")}
            </div>
          )}

          {agents.map((agent) => {
            const emoji = agentEmoji(agent);
            return (
              <button
                key={agent.agent_key}
                type="button"
                onClick={() => handleAgentSelect(agent.agent_key)}
                className="flex w-full items-center gap-3 rounded-lg border p-3 text-left hover:bg-accent transition-colors"
              >
                {emoji ? (
                  <span className="text-2xl shrink-0">{emoji}</span>
                ) : (
                  <Bot className="h-6 w-6 shrink-0 text-muted-foreground" />
                )}
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">
                    {agent.display_name || agent.agent_key}
                  </p>
                  {agent.agent_description && (
                    <p className="text-xs text-muted-foreground truncate">
                      {agent.agent_description}
                    </p>
                  )}
                </div>
                {agent.is_default && (
                  <span className="text-xs bg-primary/10 text-primary px-2 py-1 rounded whitespace-nowrap">
                    {t("default")}
                  </span>
                )}
              </button>
            );
          })}
        </div>
      </DialogContent>
    </Dialog>
  );
}
