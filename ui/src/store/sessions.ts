import { atom, getDefaultStore } from "jotai"

import type { Session } from "@/api/gateway-sessions"

export interface SessionsState {
  sessions: Session[]
  selectedSessionId: string | null
  isLoading: boolean
  error: string | null
  agentId: string | null // Track which agent's sessions these are
}

const DEFAULT_SESSIONS_STATE: SessionsState = {
  sessions: [],
  selectedSessionId: null,
  isLoading: false,
  error: null,
  agentId: null,
}

export const sessionsAtom = atom<SessionsState>(DEFAULT_SESSIONS_STATE)

// Derived atoms for convenience
export const sessionListAtom = atom((get) => get(sessionsAtom).sessions)

export const selectedSessionIdAtom = atom(
  (get) => get(sessionsAtom).selectedSessionId,
  (get, set, id: string | null) => {
    const state = get(sessionsAtom)
    set(sessionsAtom, { ...state, selectedSessionId: id })
  },
)

export const selectedSessionAtom = atom((get) => {
  const state = get(sessionsAtom)
  if (!state.selectedSessionId) return null
  return state.sessions.find((s) => s.id === state.selectedSessionId) || null
})

export const sessionsLoadingAtom = atom((get) => get(sessionsAtom).isLoading)

export const sessionsErrorAtom = atom((get) => get(sessionsAtom).error)

export const sessionsAgentIdAtom = atom((get) => get(sessionsAtom).agentId)

const store = getDefaultStore()

export function getSessionsState() {
  return store.get(sessionsAtom)
}

export function updateSessionsStore(
  patch:
    | Partial<SessionsState>
    | ((prev: SessionsState) => Partial<SessionsState> | SessionsState),
) {
  store.set(sessionsAtom, (prev) => {
    const nextPatch = typeof patch === "function" ? patch(prev) : patch
    return { ...prev, ...nextPatch }
  })
}
