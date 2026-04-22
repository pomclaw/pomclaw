import { atom, getDefaultStore } from "jotai"

import type { Agent } from "@/api/gateway-agents"

export interface AgentsState {
  agents: Agent[]
  selectedAgentId: string | null
  isLoading: boolean
  error: string | null
}

const DEFAULT_AGENTS_STATE: AgentsState = {
  agents: [],
  selectedAgentId: null,
  isLoading: false,
  error: null,
}

export const agentsAtom = atom<AgentsState>(DEFAULT_AGENTS_STATE)

// Derived atoms for convenience
export const agentListAtom = atom((get) => get(agentsAtom).agents)

export const selectedAgentIdAtom = atom(
  (get) => get(agentsAtom).selectedAgentId,
  (get, set, id: string | null) => {
    const state = get(agentsAtom)
    set(agentsAtom, { ...state, selectedAgentId: id })
  },
)

export const selectedAgentAtom = atom((get) => {
  const state = get(agentsAtom)
  if (!state.selectedAgentId) return null
  return state.agents.find((a) => a.id === state.selectedAgentId) || null
})

export const agentsLoadingAtom = atom((get) => get(agentsAtom).isLoading)

export const agentsErrorAtom = atom((get) => get(agentsAtom).error)

const store = getDefaultStore()

export function getAgentsState() {
  return store.get(agentsAtom)
}

export function updateAgentsStore(
  patch:
    | Partial<AgentsState>
    | ((prev: AgentsState) => Partial<AgentsState> | AgentsState),
) {
  store.set(agentsAtom, (prev) => {
    const nextPatch = typeof patch === "function" ? patch(prev) : patch
    return { ...prev, ...nextPatch }
  })
}
