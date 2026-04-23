import { IconPlus, IconTrash } from "@tabler/icons-react"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { useSessions } from "@/hooks/use-sessions"

export interface SessionListProps {
  agentId: string
  onCreateClick?: () => void
  onSelectSession?: (sessionId: string) => void
}

export function SessionList({
  agentId,
  onCreateClick,
  onSelectSession,
}: SessionListProps) {
  const { t } = useTranslation()
  const { sessions, selectedSession, loadSessions, deleteSession, selectSession } =
    useSessions()

  // Load sessions on mount
  useEffect(() => {
    loadSessions(agentId, 0, 20)
  }, [agentId, loadSessions])

  const handleSelectSession = (sessionId: string) => {
    selectSession(sessionId)
    onSelectSession?.(sessionId)
  }

  const handleDeleteSession = async (sessionId: string, e: React.MouseEvent) => {
    e.stopPropagation()
    if (!confirm(t("sessions.deleteConfirm", "Delete this session?"))) {
      return
    }
    try {
      await deleteSession(agentId, sessionId)
    } catch (err) {
      console.error("Failed to delete session:", err)
    }
  }

  return (
    <div className="flex h-full flex-col gap-2 p-2">
      {/* Header with New Session button */}
      <div className="flex items-center justify-between gap-2">
        <h3 className="text-sm font-medium">{t("sessions.title", "Sessions")}</h3>
        <Button
          variant="ghost"
          size="sm"
          onClick={onCreateClick}
          className="h-6 w-6 p-0"
          title={t("sessions.create", "New Session")}
        >
          <IconPlus className="size-4" />
        </Button>
      </div>

      {/* Sessions list */}
      <div className="flex-1 space-y-1 overflow-y-auto">
        {!sessions || sessions.length === 0 ? (
          <div className="text-muted-foreground flex h-24 items-center justify-center text-center text-xs">
            {t("sessions.empty", "No sessions yet. Create one to get started.")}
          </div>
        ) : (
          sessions.map((session) => (
            <div
              key={session.id}
              className={`flex items-center justify-between gap-2 rounded-md px-2 py-1.5 transition-colors cursor-pointer ${
                selectedSession?.id === session.id
                  ? "bg-accent/80 text-foreground font-medium"
                  : "hover:bg-muted/60 text-muted-foreground"
              }`}
              onClick={() => handleSelectSession(session.id)}
            >
              <div className="min-w-0 flex-1 truncate">
                <p className="text-sm truncate">{session.title || "Untitled Session"}</p>
                {session.message_count ? (
                  <p className="text-xs opacity-60">
                    {t("sessions.messageCount", "{{count}} messages", {
                      count: session.message_count,
                    })}
                  </p>
                ) : null}
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => handleDeleteSession(session.id, e)}
                className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100"
                title={t("sessions.delete", "Delete")}
              >
                <IconTrash className="size-3.5" />
              </Button>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
