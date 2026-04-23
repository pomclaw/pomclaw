import { IconPlus } from "@tabler/icons-react"
import { type ChangeEvent, useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { AssistantMessage } from "@/components/chat/assistant-message"
import {
  ChatComposer,
  type ChatInputDisabledReason,
} from "@/components/chat/chat-composer"
import { ChatEmptyState } from "@/components/chat/chat-empty-state"
import { SessionHistoryMenu } from "@/components/chat/session-history-menu"
import { TypingIndicator } from "@/components/chat/typing-indicator"
import { UserMessage } from "@/components/chat/user-message"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  connectChat,
  disconnectChat,
} from "@/features/chat/controller"
import { usePicoChat } from "@/hooks/use-pico-chat"
import { useSessionHistory } from "@/hooks/use-session-history"
import { useAgents } from "@/hooks/use-agents"
import { useSessions } from "@/hooks/use-sessions"
import { updateChatStore } from "@/store/chat"
import type { ConnectionState } from "@/store/chat"
import type { ChatAttachment } from "@/store/chat"

const MAX_IMAGE_SIZE_BYTES = 7 * 1024 * 1024
const MAX_IMAGE_SIZE_LABEL = "7 MB"
const ALLOWED_IMAGE_TYPES = new Set([
  "image/jpeg",
  "image/png",
  "image/gif",
  "image/webp",
  "image/bmp",
])

function readFileAsDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      if (typeof reader.result === "string") {
        resolve(reader.result)
        return
      }
      reject(new Error("Failed to read file"))
    }
    reader.onerror = () =>
      reject(reader.error || new Error("Failed to read file"))
    reader.readAsDataURL(file)
  })
}

function resolveChatInputDisabledReason({
  connectionState,
}: {
  connectionState: ConnectionState
}): ChatInputDisabledReason | null {
  if (connectionState === "error") {
    return "websocketError"
  }

  if (connectionState === "disconnected") {
    return "websocketDisconnected"
  }

  return null
}

export function ChatPage() {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isAtBottom, setIsAtBottom] = useState(true)
  const [hasScrolled, setHasScrolled] = useState(false)
  const [input, setInput] = useState("")
  const [attachments, setAttachments] = useState<ChatAttachment[]>([])

  const { selectedAgent } = useAgents()
  const { selectedSession, selectSession } = useSessions()

  const {
    messages,
    connectionState,
    isTyping,
    activeSessionId,
    sendMessage,
    switchSession: baseSwitchSession,
    newChat,
  } = usePicoChat()

  // Wrap switchSession to pass agentId explicitly
  const switchSession = useCallback(
    (sessionId: string) => {
      if (selectedAgent) {
        return baseSwitchSession(sessionId, selectedAgent.id)
      }
      return Promise.resolve()
    },
    [selectedAgent, baseSwitchSession],
  )

  // Update chat store when agent or session changes
  useEffect(() => {
    if (selectedAgent && selectedSession) {
      updateChatStore({
        agentId: selectedAgent.id,
        sessionId: selectedSession.id,
        activeSessionId: selectedSession.id,
      })
    }
  }, [selectedAgent, selectedSession])

  // Reconnect WebSocket when agent changes (to include agent_id in URL)
  useEffect(() => {
    if (selectedAgent?.id) {
      // Disconnect old connection and establish new one with updated agent_id
      disconnectChat()
      void connectChat()
    } else {
      disconnectChat()
    }
  }, [selectedAgent?.id])

  // Sync sessions store when activeSessionId changes (e.g., from switching via history menu)
  useEffect(() => {
    if (activeSessionId && activeSessionId !== selectedSession?.id) {
      selectSession(activeSessionId)
    }
  }, [activeSessionId, selectedSession?.id, selectSession])

  const inputDisabledReason = !activeSessionId
    ? "noSession"
    : resolveChatInputDisabledReason({
        connectionState,
      })
  const canInput = inputDisabledReason === null

  const {
    sessions,
    hasMore,
    loadError,
    loadErrorMessage,
    observerRef,
    loadSessions,
    handleDeleteSession,
  } = useSessionHistory({
    agentId: selectedAgent?.id || "",
    activeSessionId,
    onDeletedActiveSession: newChat,
  })

  const syncScrollState = (element: HTMLDivElement) => {
    const { scrollTop, scrollHeight, clientHeight } = element
    setHasScrolled(scrollTop > 0)
    setIsAtBottom(scrollHeight - scrollTop <= clientHeight + 10)
  }

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    syncScrollState(e.currentTarget)
  }

  useEffect(() => {
    if (scrollRef.current) {
      if (isAtBottom) {
        scrollRef.current.scrollTop = scrollRef.current.scrollHeight
      }
      syncScrollState(scrollRef.current)
    }
  }, [messages, isTyping, isAtBottom])

  const handleSend = () => {
    if ((!input.trim() && attachments.length === 0) || !canInput) return
    if (
      sendMessage({
        content: input,
        attachments,
      })
    ) {
      setInput("")
      setAttachments([])
    }
  }

  const handleAddImages = () => {
    if (!canInput) return
    fileInputRef.current?.click()
  }

  const handleRemoveAttachment = (index: number) => {
    setAttachments((prev) => prev.filter((_, itemIndex) => itemIndex !== index))
  }

  const handleImageSelection = async (event: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files ?? [])
    event.target.value = ""

    if (files.length === 0) {
      return
    }

    const nextAttachments: ChatAttachment[] = []
    for (const file of files) {
      if (!ALLOWED_IMAGE_TYPES.has(file.type)) {
        toast.error(
          t("chat.invalidImage", {
            name: file.name,
          }),
        )
        continue
      }

      if (file.size > MAX_IMAGE_SIZE_BYTES) {
        toast.error(
          t("chat.imageTooLarge", {
            name: file.name,
            size: MAX_IMAGE_SIZE_LABEL,
          }),
        )
        continue
      }

      try {
        nextAttachments.push({
          type: "image",
          filename: file.name,
          url: await readFileAsDataUrl(file),
        })
      } catch {
        toast.error(
          t("chat.imageReadFailed", {
            name: file.name,
          }),
        )
      }
    }

    if (nextAttachments.length > 0) {
      setAttachments(nextAttachments.slice(0, 1))
    }
  }

  const canSubmit =
    canInput && (Boolean(input.trim()) || attachments.length > 0)

  // Show selection prompt if no agent selected
  if (!selectedAgent) {
    return (
      <div className="bg-background/95 flex h-full flex-col">
        <PageHeader title={t("navigation.chat")} />
        <div className="flex flex-1 items-center justify-center p-4">
          <div className="text-center">
            <h2 className="mb-2 text-lg font-medium">
              {t("chat.selectAgent", "Select an agent to get started")}
            </h2>
            <p className="text-muted-foreground text-sm">
              {t("chat.selectAgentHint", "Choose an agent from the left sidebar")}
            </p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-background/95 flex h-full flex-col">
      <PageHeader
        title={t("navigation.chat")}
        className={`transition-shadow ${
          hasScrolled ? "shadow-xs" : "shadow-none"
        }`}
      >
        <Button
          variant="secondary"
          size="sm"
          onClick={newChat}
          className="h-9 gap-2"
        >
          <IconPlus className="size-4" />
          <span className="hidden sm:inline">{t("chat.newChat")}</span>
        </Button>

        <SessionHistoryMenu
          sessions={sessions}
          activeSessionId={activeSessionId}
          hasMore={hasMore}
          loadError={loadError}
          loadErrorMessage={loadErrorMessage}
          observerRef={observerRef}
          onOpenChange={(open) => {
            if (open) {
              void loadSessions(true)
            }
          }}
          onSwitchSession={switchSession}
          onDeleteSession={handleDeleteSession}
        />
      </PageHeader>

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="min-h-0 flex-1 overflow-y-auto px-4 py-6 md:px-8 lg:px-24 xl:px-48"
      >
        <div className="mx-auto flex w-full max-w-250 flex-col gap-8 pb-8">
          {!selectedSession && messages.length === 0 ? (
            <div className="flex flex-1 items-center justify-center py-12">
              <div className="text-center">
                <h2 className="mb-2 text-lg font-medium">
                  {t("chat.createOrSelectSession", "Create or select a session")}
                </h2>
                <p className="text-muted-foreground text-sm">
                  {t("chat.sessionHint", "Click the + button or history menu to get started")}
                </p>
              </div>
            </div>
          ) : (
            <>
              {messages.length === 0 && !isTyping && (
                <ChatEmptyState />
              )}

              {messages.map((msg) => (
                <div key={msg.id} className="flex w-full">
                  {msg.role === "assistant" ? (
                    <AssistantMessage
                      content={msg.content}
                      isThought={msg.kind === "thought"}
                      timestamp={msg.timestamp}
                    />
                  ) : (
                    <UserMessage
                      content={msg.content}
                      attachments={msg.attachments}
                    />
                  )}
                </div>
              ))}

              {isTyping && <TypingIndicator />}
            </>
          )}
        </div>
      </div>

      <input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/png,image/gif,image/webp,image/bmp"
        className="hidden"
        onChange={handleImageSelection}
      />

      <ChatComposer
        input={input}
        attachments={attachments}
        onInputChange={setInput}
        onAddImages={handleAddImages}
        onRemoveAttachment={handleRemoveAttachment}
        onSend={handleSend}
        inputDisabledReason={inputDisabledReason}
        canSend={canSubmit}
      />
    </div>
  )
}
