import { useState } from "react"

import { AgentCreateDialog } from "@/components/agents/agent-create-dialog"
import { AgentEditDialog } from "@/components/agents/agent-edit-dialog"
import { AgentList } from "@/components/agents/agent-list"
import { SessionCreateDialog } from "@/components/sessions/session-create-dialog"
import {
  Sidebar,
  SidebarContent,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar"
import { useAgents } from "@/hooks/use-agents"
import type { Agent } from "@/api/gateway-agents"

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { isMobile, setOpenMobile } = useSidebar()
  const { selectedAgent } = useAgents()

  const [showCreateAgentDialog, setShowCreateAgentDialog] = useState(false)
  const [showEditAgentDialog, setShowEditAgentDialog] = useState(false)
  const [editingAgent, setEditingAgent] = useState<Agent | null>(null)
  const [showCreateSessionDialog, setShowCreateSessionDialog] = useState(false)

  const handleCreateAgentClick = () => {
    setShowCreateAgentDialog(true)
  }

  const handleEditAgentClick = (agentId: string) => {
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
