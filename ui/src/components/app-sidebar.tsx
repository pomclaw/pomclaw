import { useState } from "react"

import { AgentCreateDialog } from "@/components/agents/agent-create-dialog"
import { AgentEditDialog } from "@/components/agents/agent-edit-dialog"
import { AgentList } from "@/components/agents/agent-list"
import { SessionCreateDialog } from "@/components/sessions/session-create-dialog"
import { SessionList } from "@/components/sessions/session-list"
import {
  Sidebar,
  SidebarContent,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar"
import { useAgents } from "@/hooks/use-agents"
import { useSessions } from "@/hooks/use-sessions"
import type { Agent } from "@/api/gateway-agents"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { isMobile, setOpenMobile } = useSidebar()
  const { selectedAgent } = useAgents()
  const { } = useSessions()

  const [showCreateAgentDialog, setShowCreateAgentDialog] = useState(false)
  const [showEditAgentDialog, setShowEditAgentDialog] = useState(false)
  const [editingAgent, setEditingAgent] = useState<Agent | null>(null)
  const [showCreateSessionDialog, setShowCreateSessionDialog] = useState(false)

  const handleCreateAgentClick = () => {
    setShowCreateAgentDialog(true)
  }

  const handleEditAgentClick = (agentId: string) => {
    // Note: In a real app, you'd fetch the agent here or use the one from state
    // For now, we'll use the selectedAgent if it matches
    const agent = selectedAgent
    if (agent && agent.id === agentId) {
      setEditingAgent(agent)
      setShowEditAgentDialog(true)
    }
  }

  const handleSelectAgent = () => {
    if (isMobile) {
      setOpenMobile(false)
    }
  }

  const handleCreateSessionClick = () => {
    setShowCreateSessionDialog(true)
  }

  const handleSelectSession = () => {
    if (isMobile) {
      setOpenMobile(false)
    }
  }

  return (
    <>
      <Sidebar
        {...props}
        className="bg-background border-r-border/20 border-r pt-3"
      >
        <SidebarContent className="bg-background flex flex-col gap-3 overflow-hidden">
          {/* Agent List */}
          <div className="flex-1 min-h-0 flex flex-col">
            <AgentList
              onCreateClick={handleCreateAgentClick}
              onEditClick={handleEditAgentClick}
              onSelectAgent={handleSelectAgent}
            />
          </div>

          {/* Session List */}
          <div className="flex-1 min-h-0 flex flex-col border-t border-border/20 pt-3">
            <SessionList
              agentId={selectedAgent?.id || null}
              onCreateClick={handleCreateSessionClick}
              onSelectSession={handleSelectSession}
            />
          </div>
        </SidebarContent>
        <SidebarRail />
      </Sidebar>

      {/* Dialogs */}
      <AgentCreateDialog
        open={showCreateAgentDialog}
        onOpenChange={setShowCreateAgentDialog}
        onSuccess={() => {
          setShowCreateAgentDialog(false)
        }}
      />

      <AgentEditDialog
        agent={editingAgent}
        open={showEditAgentDialog}
        onOpenChange={setShowEditAgentDialog}
        onSuccess={() => {
          setShowEditAgentDialog(false)
          setEditingAgent(null)
        }}
      />

      <SessionCreateDialog
        agentId={selectedAgent?.id || null}
        open={showCreateSessionDialog}
        onOpenChange={setShowCreateSessionDialog}
        onSuccess={() => {
          setShowCreateSessionDialog(false)
        }}
      />
    </>
  )
}
