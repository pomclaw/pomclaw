import {
  useEffect,
  useRef,
  useCallback,
  useMemo,
} from 'react'
import { useChat } from '@/hooks/use-chat'
import { MessageBubble } from './message-bubble'
import { ActivityIndicator } from './activity-indicator'
import { InputBar } from './input-bar'

export function ChatCanvas() {
  const { messages, isRunning, activity, sendMessage, abort } = useChat()
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const userScrolledUp = useRef(false)

  // Find last assistant message ID for streaming cursor
  const lastAssistantId = useMemo(() => {
    for (let i = messages.length - 1; i >= 0; i--) {
      if (messages[i].role === 'assistant') return messages[i].id
    }
    return null
  }, [messages])

  useEffect(() => {
    if (!userScrolledUp.current) {
      messagesEndRef.current?.scrollIntoView({
        behavior: isRunning ? 'smooth' : 'instant',
      })
    }
  }, [messages, isRunning])

  const handleScroll = useCallback(() => {
    const el = scrollAreaRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 50
    userScrolledUp.current = !atBottom
  }, [])

  const handleSend = useCallback(
    (text: string) => {
      userScrolledUp.current = false
      sendMessage({ message: text, agentId: 'default' })
    },
    [sendMessage]
  )

  const handleStop = useCallback(() => {
    abort()
  }, [abort])

  const hasMessages = messages.length > 0

  return (
    <div className="flex-1 flex flex-col min-h-0">
      {/* Chat body */}
      <div className="flex-1 flex flex-col min-h-0">
        {/* Messages area */}
        <div
          ref={scrollAreaRef}
          onScroll={handleScroll}
          className="flex-1 overflow-y-auto overscroll-contain px-4 py-2"
        >
          <div className="max-w-3xl mx-auto">
            {!hasMessages && <EmptyState onSuggestion={handleSend} />}

            {messages.map((msg) => (
              <MessageBubble
                key={msg.id}
                message={msg}
                isStreaming={isRunning && msg.id === lastAssistantId}
              />
            ))}

            {isRunning && activity && (
              <ActivityIndicator
                phase={activity.phase}
                tool={activity.tool}
                iteration={activity.iteration}
              />
            )}

            <div ref={messagesEndRef} />
          </div>
        </div>

        {/* Input bar */}
        <InputBar
          onSend={handleSend}
          onStop={handleStop}
          isRunning={isRunning}
          placeholder="Send a message..."
        />
      </div>
    </div>
  )
}

/** Empty state with logo and clickable suggested prompts */
function EmptyState({
  onSuggestion,
}: {
  onSuggestion?: (text: string) => void
}) {
  const suggestions = [
    'What can you help me with?',
    'List files in current directory',
    'Explain this codebase structure',
  ]

  return (
    <div className="flex flex-col items-center justify-center text-center py-20">
      <div className="h-14 w-14 mb-5 rounded-lg bg-primary/10 flex items-center justify-center">
        <span className="text-2xl">🤖</span>
      </div>
      <h2 className="text-lg font-medium text-foreground mb-1">
        Welcome to Pomclaw
      </h2>
      <p className="text-sm text-muted-foreground max-w-sm mb-6">
        Your personal AI assistant powered by Eino framework. Ask me anything!
      </p>
      <div className="flex flex-wrap justify-center gap-2">
        {suggestions.map((prompt) => (
          <button
            key={prompt}
            onClick={() => onSuggestion?.(prompt)}
            className="text-xs text-muted-foreground bg-secondary border border-border rounded-full px-3 py-1.5 hover:border-primary/40 hover:text-primary transition-colors cursor-pointer"
          >
            {prompt}
          </button>
        ))}
      </div>
    </div>
  )
}
