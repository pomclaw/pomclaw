import type { ChatMessage } from '@/store/chat-messages'
import { ThinkingBlock } from './thinking-block'
import { ToolCallBlock } from './tool-call-block'
import { MarkdownRenderer } from './markdown-renderer'

interface MessageBubbleProps {
  message: ChatMessage
  /** True when this is the last assistant message during an active run */
  isStreaming?: boolean
}

function formatTimestamp(timestamp: number): string {
  const date = new Date(timestamp)
  const now = new Date()
  const isToday = date.toDateString() === now.toDateString()

  if (isToday) {
    return date.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
    })
  }

  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function MessageBubble({ message, isStreaming }: MessageBubbleProps) {
  if (message.role === 'user') {
    return (
      <div className="flex justify-end mb-4">
        <div className="max-w-[85%] bg-secondary border border-border shadow-sm border-r-2 border-r-primary rounded-xl px-4 py-3">
          <p className="text-sm text-foreground whitespace-pre-wrap break-words">
            {message.content}
          </p>
          <time className="text-[10px] text-muted-foreground mt-1 block text-right">
            {formatTimestamp(message.timestamp)}
          </time>
        </div>
      </div>
    )
  }

  // Assistant message
  const outputTokens = message.usage?.outputTokens ?? 0
  const toolCalls = message.toolCalls ?? []
  const hasTools = toolCalls.length > 0

  return (
    <div className="mb-6 min-w-0">
      {message.thinkingText && (
        <ThinkingBlock
          text={message.thinkingText}
          isStreaming={isStreaming && !message.content}
        />
      )}

      {/* Group tool calls in a single bordered container */}
      {hasTools && (
        <div className="mb-2 rounded-md border border-border bg-muted/30 divide-y divide-border overflow-hidden">
          {toolCalls.map((tc) => (
            <ToolCallBlock key={tc.toolId} toolCall={tc} compact />
          ))}
        </div>
      )}

      {message.content && (
        <div className="min-w-0 overflow-hidden">
          <MarkdownRenderer content={message.content} />
          {/* Streaming cursor on last streaming message */}
          {isStreaming && (
            <span className="inline-block w-0.5 h-4 bg-primary animate-pulse rounded-sm ml-0.5 align-text-bottom" />
          )}
        </div>
      )}

      {/* Streaming cursor when no content yet but run is active */}
      {isStreaming &&
        !message.content &&
        !message.thinkingText &&
        toolCalls.length === 0 && (
          <span className="inline-block w-0.5 h-4 bg-primary animate-pulse rounded-sm" />
        )}

      <div className="flex items-center gap-2 mt-1.5 text-[10px] text-muted-foreground">
        <time>{formatTimestamp(message.timestamp)}</time>
        {outputTokens > 0 && (
          <span>· {outputTokens.toLocaleString()} tokens</span>
        )}
      </div>
    </div>
  )
}
