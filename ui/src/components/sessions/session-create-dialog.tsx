import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useSessions } from "@/hooks/use-sessions"

export interface SessionCreateDialogProps {
  agentId: string | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function SessionCreateDialog({
  agentId,
  open,
  onOpenChange,
  onSuccess,
}: SessionCreateDialogProps) {
  const { t } = useTranslation()
  const { createSession, isLoading, error } = useSessions()

  const [title, setTitle] = useState("")
  const [localError, setLocalError] = useState("")

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError("")

    if (!agentId) {
      setLocalError(t("sessions.errors.agentRequired", "Agent is required"))
      return
    }

    try {
      await createSession(agentId, title.trim() || undefined)

      // Reset form and close
      setTitle("")
      onOpenChange(false)
      onSuccess?.()
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to create session"
      setLocalError(msg)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{t("sessions.create", "New Session")}</DialogTitle>
          <DialogDescription>
            {t("sessions.createDescription", "Create a new chat session")}
          </DialogDescription>
        </DialogHeader>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="session-title">
              {t("sessions.title", "Session Title")} ({t("labels.optional", "optional")})
            </Label>
            <Input
              id="session-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t("sessions.titlePlaceholder", "e.g., Project Discussion")}
              disabled={isLoading}
              autoFocus
            />
          </div>

          {error || localError ? (
            <p className="text-destructive text-sm" role="alert">
              {error || localError}
            </p>
          ) : null}

          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              {t("labels.cancel", "Cancel")}
            </Button>
            <Button type="submit" disabled={isLoading || !agentId}>
              {isLoading
                ? t("labels.creating", "Creating...")
                : t("sessions.create", "Create Session")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
