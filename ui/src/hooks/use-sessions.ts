import { useAtom, useAtomValue } from "jotai"
import { useCallback, useState } from "react"

import {
  createSession,
  deleteSession,
  getSession,
  listSessions,
  loadSessionChatHistory,
  updateSession,
  type Session,
} from "@/api/gateway-sessions"
import {
  sessionsErrorAtom,
  sessionsLoadingAtom,
  sessionsAtom,
  selectedSessionAtom,
  selectedSessionIdAtom,
  updateSessionsStore,
} from "@/store/sessions"

export function useSessions() {
  const [sessionsState] = useAtom(sessionsAtom)
  const isLoading = useAtomValue(sessionsLoadingAtom)
  const error = useAtomValue(sessionsErrorAtom)
  const selectedSession = useAtomValue(selectedSessionAtom)
  const [selectedSessionId, setSelectedSessionId] = useAtom(selectedSessionIdAtom)

  const [internalError, setInternalError] = useState<string | null>(null)

  const loadSessions = useCallback(async (agentId: string) => {
    setInternalError(null)
    updateSessionsStore({ isLoading: true, error: null, agentId })

    try {
      const response = await listSessions(agentId)
      updateSessionsStore({
        sessions: response.sessions,
        agentId,
        isLoading: false,
        selectedSessionId: null, // Reset selection when loading new agent's sessions
      })
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to load sessions"
      updateSessionsStore({
        isLoading: false,
        error: msg,
      })
      setInternalError(msg)
    }
  }, [])

  const createNewSession = useCallback(
    async (agentId: string, title?: string) => {
      setInternalError(null)
      updateSessionsStore({ isLoading: true, error: null })

      try {
        const newSession = await createSession(agentId, title)
        updateSessionsStore((prev) => ({
          ...prev,
          sessions: [...(prev.sessions || []), newSession],
          isLoading: false,
          selectedSessionId: newSession.id,
        }))
        return newSession
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to create session"
        updateSessionsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  const updateExistingSession = useCallback(
    async (sessionId: string, updates: Partial<Session>) => {
      setInternalError(null)
      updateSessionsStore({ isLoading: true, error: null })

      try {
        const updated = await updateSession(sessionId, updates)
        updateSessionsStore((prev) => ({
          ...prev,
          sessions: prev.sessions.map((s) => (s.id === sessionId ? updated : s)),
          isLoading: false,
        }))
        return updated
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to update session"
        updateSessionsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  const deleteExistingSession = useCallback(
    async (sessionId: string) => {
      setInternalError(null)
      updateSessionsStore({ isLoading: true, error: null })

      try {
        await deleteSession(sessionId)
        updateSessionsStore((prev) => ({
          ...prev,
          sessions: (prev.sessions || []).filter((s) => s.id !== sessionId),
          selectedSessionId:
            prev.selectedSessionId === sessionId ? null : prev.selectedSessionId,
          isLoading: false,
        }))
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to delete session"
        updateSessionsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  const selectSession = useCallback((sessionId: string | null) => {
    setSelectedSessionId(sessionId)
  }, [setSelectedSessionId])

  const refreshSession = useCallback(
    async (sessionId: string) => {
      try {
        const session = await getSession(sessionId)
        updateSessionsStore((prev) => ({
          ...prev,
          sessions: prev.sessions.map((s) => (s.id === sessionId ? session : s)),
        }))
        return session
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to refresh session"
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  const getChatHistory = useCallback(
    async (sessionId: string) => {
      try {
        return await loadSessionChatHistory(sessionId)
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to load chat history"
        setInternalError(msg)
        throw err
      }
    },
    [],
  )

  return {
    sessions: sessionsState.sessions,
    selectedSession,
    selectedSessionId,
    isLoading,
    error: error || internalError,
    loadSessions,
    createSession: createNewSession,
    updateSession: updateExistingSession,
    deleteSession: deleteExistingSession,
    selectSession,
    refreshSession,
    getChatHistory,
  }
}
