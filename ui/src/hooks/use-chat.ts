import { useEffect, useRef, useCallback } from 'react'
import { useAtom, useSetAtom } from 'jotai'
import { getWsClient } from '@/lib/ws-client'
import { useStreamBatcher } from './use-stream-batcher'
import {
  messagesAtom,
  addUserMessageAtom,
  addAssistantMessageAtom,
  appendChunkAtom,
  appendThinkingAtom,
  addToolCallAtom,
  updateToolResultAtom,
  finalizeMessageAtom,
  appendErrorToLastMessageAtom,
  setMessagesAtom,
  clearMessagesAtom,
  type ChatMessage,
} from '@/store/chat-messages'
import {
  isRunningAtom,
  activityAtom,
  startRunAtom,
  setActivityAtom,
  completeRunAtom,
  failRunAtom,
  cancelRunAtom,
} from '@/store/chat-activity'

interface SendMessageParams {
  message: string
  agentId?: string
  sessionKey?: string
}

export function useChat() {
  const [messages] = useAtom(messagesAtom)
  const [isRunning] = useAtom(isRunningAtom)
  const [activity] = useAtom(activityAtom)

  const addUserMessage = useSetAtom(addUserMessageAtom)
  const addAssistantMessage = useSetAtom(addAssistantMessageAtom)
  const appendChunk = useSetAtom(appendChunkAtom)
  const appendThinking = useSetAtom(appendThinkingAtom)
  const addToolCall = useSetAtom(addToolCallAtom)
  const updateToolResult = useSetAtom(updateToolResultAtom)
  const finalizeMessage = useSetAtom(finalizeMessageAtom)
  const appendErrorToLastMessage = useSetAtom(appendErrorToLastMessageAtom)
  const setMessages = useSetAtom(setMessagesAtom)
  const clearMessages = useSetAtom(clearMessagesAtom)

  const startRun = useSetAtom(startRunAtom)
  const setActivity = useSetAtom(setActivityAtom)
  const completeRun = useSetAtom(completeRunAtom)
  const failRun = useSetAtom(failRunAtom)
  const cancelRun = useSetAtom(cancelRunAtom)

  const currentSessionKeyRef = useRef<string | null>(null)
  const currentRunIdRef = useRef<string | null>(null)

  const chunkBatcher = useStreamBatcher(
    useCallback((text: string) => appendChunk(text), [appendChunk])
  )
  const thinkingBatcher = useStreamBatcher(
    useCallback((text: string) => appendThinking(text), [appendThinking])
  )

  // Subscribe to agent events from WS
  useEffect(() => {
    let ws: ReturnType<typeof getWsClient> | null = null
    try {
      ws = getWsClient()
    } catch {
      return // Not initialized yet
    }

    const unsub = ws.on('agent', (raw: unknown) => {
      const event = raw as {
        type: string
        runId?: string
        sessionKey?: string
        content?: string
        // Tool call fields
        id?: string
        name?: string
        arguments?: Record<string, unknown>
        result?: string
        is_error?: boolean
        // Activity fields
        phase?: string
        tool?: string
        iteration?: number
        // Usage fields
        usage?: {
          prompt_tokens?: number
          completion_tokens?: number
        }
        error?: string
        attempt?: number
      }

      console.log('[useChat] Received event:', event.type, {
        sessionKey: event.sessionKey,
        currentSessionKey: currentSessionKeyRef.current,
        content: event.content?.substring(0, 50),
      })

      // Filter events by session key
      if (
        currentSessionKeyRef.current &&
        event.sessionKey &&
        event.sessionKey !== currentSessionKeyRef.current
      ) {
        console.warn('[useChat] Event filtered - session key mismatch')
        return
      }

      switch (event.type) {
        case 'run.started':
          currentRunIdRef.current = event.runId ?? ''
          addAssistantMessage(event.runId ?? '')
          startRun(event.runId ?? '')
          break

        case 'chunk':
          chunkBatcher.append(event.content ?? '')
          break

        case 'thinking':
          thinkingBatcher.append(event.content ?? '')
          break

        case 'tool.call':
          chunkBatcher.flush()
          addToolCall({
            toolId: event.id ?? '',
            toolName: event.name ?? 'unknown',
            arguments: event.arguments ?? {},
          })
          break

        case 'tool.result': {
          const isError = event.is_error ?? false
          updateToolResult({
            toolId: event.id ?? '',
            result: isError ? '' : (event.result ?? ''),
            error: isError
              ? (event.content ?? event.result ?? 'Error')
              : undefined,
          })
          break
        }

        case 'activity':
          setActivity({
            phase: event.phase ?? 'thinking',
            tool: event.tool,
            iteration: event.iteration,
          })
          break

        case 'run.completed': {
          chunkBatcher.flush()
          thinkingBatcher.flush()
          finalizeMessage({
            content: event.content ?? '',
            usage: event.usage
              ? {
                  inputTokens: event.usage.prompt_tokens ?? 0,
                  outputTokens: event.usage.completion_tokens ?? 0,
                }
              : undefined,
          })
          completeRun()
          currentRunIdRef.current = null
          break
        }

        case 'run.failed':
          chunkBatcher.flush()
          thinkingBatcher.flush()
          appendErrorToLastMessage(event.error ?? 'Unknown error')
          failRun()
          currentRunIdRef.current = null
          break

        case 'run.cancelled':
          chunkBatcher.flush()
          thinkingBatcher.flush()
          cancelRun()
          currentRunIdRef.current = null
          break

        case 'run.retrying':
          setActivity({
            phase: 'retrying',
            tool: undefined,
            iteration: Number(event.attempt) || 0,
          })
          break
      }
    })

    return unsub
  }, [
    addAssistantMessage,
    startRun,
    appendChunk,
    appendThinking,
    addToolCall,
    updateToolResult,
    setActivity,
    finalizeMessage,
    appendErrorToLastMessage,
    completeRun,
    failRun,
    cancelRun,
    chunkBatcher,
    thinkingBatcher,
  ])

  const sendMessage = useCallback(
    async ({ message, agentId = 'default', sessionKey }: SendMessageParams) => {
      let ws: ReturnType<typeof getWsClient>
      try {
        ws = getWsClient()
      } catch (err) {
        console.error('[useChat] WsClient not initialized:', err)
        return
      }

      if (!ws.isConnected) {
        console.warn('[useChat] WebSocket not connected yet')
        return
      }

      if (!message.trim()) return

      // Generate session key if not provided
      let sk = sessionKey
      if (!sk) {
        sk = `agent:${agentId}:ws:direct:system:${crypto.randomUUID().slice(0, 8)}`
      }

      currentSessionKeyRef.current = sk

      // Add user message to UI
      addUserMessage(message)

      // Send chat.send RPC
      try {
        await ws.call('chat.send', {
          message,
          agentId,
          sessionKey: sk,
          stream: true,
        })
      } catch (err) {
        console.error('[useChat] chat.send failed:', err)
        appendErrorToLastMessage('Failed to send message')
      }
    },
    [addUserMessage, appendErrorToLastMessage]
  )

  const loadHistory = useCallback(
    async (sessionKey: string) => {
      let ws: ReturnType<typeof getWsClient>
      try {
        ws = getWsClient()
      } catch {
        return
      }

      try {
        const result = (await ws.call('chat.history', {
          sessionKey,
        })) as { messages?: ChatMessage[] }

        if (result?.messages) {
          setMessages(result.messages)
        }
      } catch (err) {
        console.error('[useChat] Failed to load history:', err)
      }
    },
    [setMessages]
  )

  const abort = useCallback(async () => {
    const sk = currentSessionKeyRef.current
    const runId = currentRunIdRef.current
    if (!sk && !runId) return

    let ws: ReturnType<typeof getWsClient>
    try {
      ws = getWsClient()
    } catch {
      return
    }

    try {
      await ws.call('chat.abort', { sessionKey: sk, runId })
    } catch {
      // Ignore errors
    }
  }, [])

  return {
    messages,
    isRunning,
    activity,
    sendMessage,
    loadHistory,
    abort,
    clearMessages,
  }
}
