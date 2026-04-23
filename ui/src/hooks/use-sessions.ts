import { useAtom, useAtomValue } from "jotai"
import { useCallback, useState } from "react"

import {
  createSession,
  deleteSession,
  getSessionHistory,
  getSessions,
  type SessionDetail,
} from "@/api/sessions"
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

  const loadSessions = useCallback(
    async (agentId: string, offset: number = 0, limit: number = 20) => {
      setInternalError(null)
      updateSessionsStore({ isLoading: true, error: null })

      try {
        const sessions = await getSessions(agentId, offset, limit)
        updateSessionsStore({
          sessions,
          isLoading: false,
          selectedSessionId: null,
        })
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to load sessions"
        updateSessionsStore({
          isLoading: false,
          error: msg,
        })
        setInternalError(msg)
      }
    },
    [],
  )

  const createNewSession = useCallback(
    async (agentId: string, title?: string) => {
      setInternalError(null)
      updateSessionsStore({ isLoading: true, error: null })

      try {
        const newSession = await createSession(agentId, title)
        const listItem = {
          id: newSession.id,
          title: title || "",
          preview: title || "(empty)",
          message_count: 0,
          created: newSession.created,
          updated: newSession.updated,
        }
        updateSessionsStore((prev) => ({
          ...prev,
          sessions: [...(prev.sessions || []), listItem],
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

  const deleteExistingSession = useCallback(
    async (agentId: string, sessionId: string) => {
      setInternalError(null)
      updateSessionsStore({ isLoading: true, error: null })

      try {
        await deleteSession(agentId, sessionId)
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

  const getSessionDetail = useCallback(
    async (agentId: string, sessionId: string): Promise<SessionDetail> => {
      try {
        return await getSessionHistory(agentId, sessionId)
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to get session"
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
    deleteSession: deleteExistingSession,
    selectSession,
    getSessionDetail,
  }
}
