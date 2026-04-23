import { useAtom, useAtomValue } from "jotai"
import { useCallback, useState } from "react"

import {
  createAgent,
  deleteAgent,
  getAgent,
  listAgents,
  updateAgent,
  type Agent,
} from "@/api/gateway-agents"
import {
  agentsErrorAtom,
  agentsLoadingAtom,
  agentsAtom,
  selectedAgentAtom,
  selectedAgentIdAtom,
  updateAgentsStore,
} from "@/store/agents"

export function useAgents() {
  const [agentsState] = useAtom(agentsAtom)
  const isLoading = useAtomValue(agentsLoadingAtom)
  const error = useAtomValue(agentsErrorAtom)
  const selectedAgent = useAtomValue(selectedAgentAtom)
  const [selectedAgentId, setSelectedAgentId] = useAtom(selectedAgentIdAtom)

  const [internalError, setInternalError] = useState<string | null>(null)

  const loadAgents = useCallback(async () => {
    setInternalError(null)
    updateAgentsStore({ isLoading: true, error: null })

    try {
      const response = await listAgents()
      updateAgentsStore({
        agents: response.agents,
        isLoading: false,
      })
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to load agents"
      updateAgentsStore({
        isLoading: false,
        error: msg,
      })
      setInternalError(msg)
    }
  }, [])

  const createNewAgent = useCallback(
    async (agent: Partial<Agent>) => {
      setInternalError(null)
      updateAgentsStore({ isLoading: true, error: null })

      try {
        const newAgent = await createAgent(agent)
        updateAgentsStore((prev) => ({
          ...prev,
          agents: [...(prev.agents || []), newAgent],
          isLoading: false,
        }))
        // 创建成功后，重新加载所有 agents 确保列表最新
        await loadAgents()
        return newAgent
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to create agent"
        updateAgentsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        throw err
      }
    },
    [loadAgents],
  )

  const updateExistingAgent = useCallback(
    async (agentId: string, updates: Partial<Agent>) => {
      setInternalError(null)
      updateAgentsStore({ isLoading: true, error: null })

      try {
        const updated = await updateAgent(agentId, updates)
        updateAgentsStore((prev) => ({
          ...prev,
          agents: (prev.agents || []).map((a) => (a.id === agentId ? updated : a)),
          isLoading: false,
        }))
        return updated
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to update agent"
        updateAgentsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  const deleteExistingAgent = useCallback(
    async (agentId: string) => {
      setInternalError(null)
      updateAgentsStore({ isLoading: true, error: null })

      try {
        await deleteAgent(agentId)
        updateAgentsStore((prev) => ({
          ...prev,
          agents: (prev.agents || []).filter((a) => a.id !== agentId),
          selectedAgentId:
            prev.selectedAgentId === agentId ? null : prev.selectedAgentId,
          isLoading: false,
        }))
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to delete agent"
        updateAgentsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  const selectAgent = useCallback((agentId: string | null) => {
    setSelectedAgentId(agentId)
  }, [setSelectedAgentId])

  const refreshAgent = useCallback(
    async (agentId: string) => {
      try {
        const agent = await getAgent(agentId)
        updateAgentsStore((prev) => ({
          ...prev,
          agents: prev.agents.map((a) => (a.id === agentId ? agent : a)),
        }))
        return agent
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to refresh agent"
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  return {
    agents: agentsState.agents,
    selectedAgent,
    selectedAgentId,
    isLoading,
    error: error || internalError,
    loadAgents,
    createAgent: createNewAgent,
    updateAgent: updateExistingAgent,
    deleteAgent: deleteExistingAgent,
    selectAgent,
    refreshAgent,
  }
}
