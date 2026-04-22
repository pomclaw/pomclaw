import { IconPlus, IconTrash, IconEdit } from "@tabler/icons-react"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { useAgents } from "@/hooks/use-agents"

export interface AgentListProps {
  onCreateClick?: () => void
  onEditClick?: (agentId: string) => void
  onSelectAgent?: (agentId: string) => void
}

export function AgentList({
  onCreateClick,
  onEditClick,
  onSelectAgent,
}: AgentListProps) {
  const { t } = useTranslation()
  const { agents, selectedAgent, loadAgents, deleteAgent, selectAgent } = useAgents()

  // Load agents on mount
  useEffect(() => {
    loadAgents()
  }, [loadAgents])

  const handleSelectAgent = (agentId: string) => {
    selectAgent(agentId)
    onSelectAgent?.(agentId)
  }

  const handleDeleteAgent = async (agentId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm(t("agents.deleteConfirm", "Are you sure you want to delete this agent?"))) {
      return
    }
    try {
      await deleteAgent(agentId)
    } catch (err) {
      console.error("Failed to delete agent:", err)
    }
  }

  return (
    <div className="flex h-full flex-col gap-2 p-2">
      {/* Header with New Agent button */}
      <div className="flex items-center justify-between gap-2">
        <h3 className="text-sm font-medium">{t("agents.title", "Agents")}</h3>
        <Button
          variant="ghost"
          size="sm"
          onClick={onCreateClick}
          className="h-6 w-6 p-0"
          title={t("agents.create", "New Agent")}
        >
          <IconPlus className="size-4" />
        </Button>
      </div>

      {/* Agents list */}
      <div className="flex-1 space-y-1 overflow-y-auto">
        {!agents || agents.length === 0 ? (
          <div className="text-muted-foreground flex h-24 items-center justifyfycenter text-center text-xs">
            {t("agents.empty", "No agents yet. Create one to get started.")}
          </div>
        ) : (
          agents.map((agent) => (
            <div
              key={agent.id}
              className={`flex items-center justify-between gap-2 rounded-md px-2 py-1.5 transition-colors cursor-pointer ${
                selectedAgent?.id === agent.id
                  ? "bg-accent/80 text-foreground font-medium"
                  : "hover:bg-muted/60 text-muted-foreground"
              }`}
              onClick={() => handleSelectAgent(agent.id)}
            >
              <div className="min-w-0 flex-1 truncate">
                <p className="text-sm truncate">{agent.name}</p>
                {agent.description && (
                  <p className="text-xs opacity-60 truncate">{agent.description}</p>
                )}
              </div>
              <div className="flex gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation()
                    onEditClick?.(agent.id)
                  }}
                  className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
                  title={t("agents.edit", "Edit")}
                >
                  <IconEdit className="size-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={(e) => handleDeleteAgent(agent.id, e)}
                  className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
                  title={t("agents.delete", "Delete")}
                >
                  <IconTrash className="size-3.5" />
                </Button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
