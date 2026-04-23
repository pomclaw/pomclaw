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
import { Textarea } from "@/components/ui/textarea"
import { useAgents } from "@/hooks/use-agents"

export interface AgentCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function AgentCreateDialog({
  open,
  onOpenChange,
  onSuccess,
}: AgentCreateDialogProps) {
  const { t } = useTranslation()
  const { createAgent, isLoading, error } = useAgents()

  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [systemPrompt, setSystemPrompt] = useState("")
  const [model, setModel] = useState("claude-opus-4-6")
  const [localError, setLocalError] = useState("")

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLocalError("")

    if (!name.trim()) {
      setLocalError(t("agents.errors.nameRequired", "Agent name is required"))
      return
    }

    try {
      await createAgent({
        name: name.trim(),
        description: description.trim() || undefined,
        system_prompt: systemPrompt.trim() || undefined,
        model: model.trim() || undefined,
      })

      // Reset form and close
      setName("")
      setDescription("")
      setSystemPrompt("")
      setModel("claude-opus-4-6")
      onOpenChange(false)
      onSuccess?.()
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Failed to create agent"
      setLocalError(msg)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{t("agents.create", "Create Agent")}</DialogTitle>
          <DialogDescription>
            {t("agents.createDescription", "Create a new AI agent with custom settings")}
          </DialogDescription>
        </DialogHeader>

        <form className="space-y-4" onSubmit={handleSubmit}>
          <div className="space-y-2">
            <Label htmlFor="agent-name">{t("agents.name", "Agent Name")}</Label>
            <Input
              id="agent-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t("agents.namePlaceholder", "Enter agent name")}
              disabled={isLoading}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="agent-description">
              {t("agents.description", "Description")}
            </Label>
            <Textarea
              id="agent-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t("agents.descriptionPlaceholder", "Brief description (optional)")}
              disabled={isLoading}
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="agent-system-prompt">
              {t("agents.systemPrompt", "System Prompt")}
            </Label>
            <Textarea
              id="agent-system-prompt"
              value={systemPrompt}
              onChange={(e) => setSystemPrompt(e.target.value)}
              placeholder={t(
                "agents.systemPromptPlaceholder",
                "Custom system prompt (optional)",
              )}
              disabled={isLoading}
              rows={4}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="agent-model">{t("agents.model", "Model")}</Label>
            <Input
              id="agent-model"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              placeholder={t("agents.modelPlaceholder", "e.g., claude-opus-4-6")}
              disabled={isLoading}
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
            <Button type="submit" disabled={isLoading}>
              {isLoading
                ? t("labels.creating", "Creating...")
                : t("agents.create", "Create Agent")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
