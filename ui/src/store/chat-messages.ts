// Chat message atoms using Jotai
// Ported from GoClaw's chat-message-store.ts (Zustand → Jotai)

import { atom } from 'jotai'

export interface ToolCall {
  toolId: string
  toolName: string
  arguments: Record<string, unknown>
  state: 'calling' | 'completed' | 'error'
  result?: string
  error?: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: number
  // Assistant-only fields
  thinkingText?: string
  toolCalls?: ToolCall[]
  media?: { type: string; url: string }[]
  usage?: { inputTokens: number; outputTokens: number }
}

// Base atom
export const messagesAtom = atom<ChatMessage[]>([])

// Setter atoms for each action
export const addUserMessageAtom = atom(
  null,
  (get, set, content: string) => {
    const newMsg: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      timestamp: Date.now(),
    }
    set(messagesAtom, [...get(messagesAtom), newMsg])
  }
)

export const addAssistantMessageAtom = atom(
  null,
  (get, set, id: string) => {
    const newMsg: ChatMessage = {
      id,
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
      toolCalls: [],
    }
    set(messagesAtom, [...get(messagesAtom), newMsg])
  }
)

export const appendChunkAtom = atom(
  null,
  (get, set, text: string) => {
    const msgs = [...get(messagesAtom)]
    const last = msgs[msgs.length - 1]
    if (last?.role === 'assistant') {
      msgs[msgs.length - 1] = { ...last, content: last.content + text }
      set(messagesAtom, msgs)
    }
  }
)

export const appendThinkingAtom = atom(
  null,
  (get, set, text: string) => {
    const msgs = [...get(messagesAtom)]
    const last = msgs[msgs.length - 1]
    if (last?.role === 'assistant') {
      msgs[msgs.length - 1] = {
        ...last,
        thinkingText: (last.thinkingText ?? '') + text,
      }
      set(messagesAtom, msgs)
    }
  }
)

export const addToolCallAtom = atom(
  null,
  (get, set, tc: Omit<ToolCall, 'state'>) => {
    const msgs = [...get(messagesAtom)]
    const last = msgs[msgs.length - 1]
    if (last?.role === 'assistant') {
      const toolCalls = [...(last.toolCalls ?? []), { ...tc, state: 'calling' as const }]
      msgs[msgs.length - 1] = { ...last, toolCalls }
      set(messagesAtom, msgs)
    }
  }
)

export const updateToolResultAtom = atom(
  null,
  (get, set, args: { toolId: string; result: string; error?: string }) => {
    const { toolId, result, error } = args
    const msgs = [...get(messagesAtom)]
    const last = msgs[msgs.length - 1]
    if (last?.role === 'assistant' && last.toolCalls) {
      const toolCalls = last.toolCalls.map((tc) =>
        tc.toolId === toolId
          ? { ...tc, state: (error ? 'error' : 'completed') as ToolCall['state'], result, error }
          : tc
      )
      msgs[msgs.length - 1] = { ...last, toolCalls }
      set(messagesAtom, msgs)
    }
  }
)

// Called when a run completes — sets final content/usage/media on last assistant message.
export const finalizeMessageAtom = atom(
  null,
  (get, set, args: { content?: string; usage?: ChatMessage['usage']; media?: ChatMessage['media'] }) => {
    const { content, usage, media } = args
    const msgs = [...get(messagesAtom)]
    const last = msgs[msgs.length - 1]
    if (last?.role === 'assistant') {
      msgs[msgs.length - 1] = {
        ...last,
        content: content || last.content,
        usage,
        media,
      }
      set(messagesAtom, msgs)
    }
  }
)

// Called when a run fails — appends error text if assistant message is empty.
export const appendErrorToLastMessageAtom = atom(
  null,
  (get, set, error: string) => {
    const msgs = [...get(messagesAtom)]
    const last = msgs[msgs.length - 1]
    if (last?.role === 'assistant') {
      msgs[msgs.length - 1] = { ...last, content: last.content || `Error: ${error}` }
      set(messagesAtom, msgs)
    }
  }
)

export const setMessagesAtom = atom(
  null,
  (_get, set, messages: ChatMessage[]) => {
    set(messagesAtom, messages)
  }
)

export const clearMessagesAtom = atom(
  null,
  (_get, set) => {
    set(messagesAtom, [])
  }
)
