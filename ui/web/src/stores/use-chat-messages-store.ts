import { create } from "zustand";
import type { ChatMessage } from "@/types/chat";

const MAX_CACHED_SESSIONS = 25;

interface SessionMessages {
  messages: ChatMessage[];
  streamText: string | null;
  thinkingText: string | null;
  isRunning: boolean;
  lastAccessedAt: number;
}

interface ChatMessagesState {
  sessions: Record<string, SessionMessages>;
  setSessionMessages: (sessionId: string, messages: ChatMessage[]) => void;
  updateSessionMessages: (sessionId: string, updater: (prev: ChatMessage[]) => ChatMessage[]) => void;
  setSessionStream: (sessionId: string, streamText: string | null) => void;
  setSessionThinking: (sessionId: string, thinkingText: string | null) => void;
  setSessionRunning: (sessionId: string, isRunning: boolean) => void;
}

// Evict oldest idle sessions when cache exceeds limit.
// Never evict sessions with an active run (isRunning).
function evictStale(sessions: Record<string, SessionMessages>): Record<string, SessionMessages> {
  const keys = Object.keys(sessions);
  if (keys.length <= MAX_CACHED_SESSIONS) return sessions;

  const evictable = keys
    .filter((k) => !sessions[k]?.isRunning)
    .sort((a, b) => (sessions[a]?.lastAccessedAt ?? 0) - (sessions[b]?.lastAccessedAt ?? 0));

  const toRemove = evictable.slice(0, keys.length - MAX_CACHED_SESSIONS);
  if (toRemove.length === 0) return sessions;

  const next = { ...sessions };
  for (const k of toRemove) delete next[k];
  return next;
}

function touchSession(existing: SessionMessages | undefined, patch: Partial<SessionMessages>): SessionMessages {
  return {
    messages: patch.messages ?? existing?.messages ?? [],
    streamText: patch.streamText !== undefined ? patch.streamText : (existing?.streamText ?? null),
    thinkingText: patch.thinkingText !== undefined ? patch.thinkingText : (existing?.thinkingText ?? null),
    isRunning: patch.isRunning !== undefined ? patch.isRunning : (existing?.isRunning ?? false),
    lastAccessedAt: Date.now(),
  };
}

export const useChatMessagesStore = create<ChatMessagesState>((set) => ({
  sessions: {},

  setSessionMessages: (sessionId, messages) => {
    set((state) => ({
      sessions: evictStale({
        ...state.sessions,
        [sessionId]: touchSession(state.sessions[sessionId], { messages }),
      }),
    }));
  },

  updateSessionMessages: (sessionId, updater) => {
    set((state) => {
      const current = state.sessions[sessionId]?.messages ?? [];
      return {
        sessions: evictStale({
          ...state.sessions,
          [sessionId]: touchSession(state.sessions[sessionId], { messages: updater(current) }),
        }),
      };
    });
  },

  setSessionStream: (sessionId, streamText) => {
    set((state) => ({
      sessions: {
        ...state.sessions,
        [sessionId]: touchSession(state.sessions[sessionId], { streamText }),
      },
    }));
  },

  setSessionThinking: (sessionId, thinkingText) => {
    set((state) => ({
      sessions: {
        ...state.sessions,
        [sessionId]: touchSession(state.sessions[sessionId], { thinkingText }),
      },
    }));
  },

  setSessionRunning: (sessionId, isRunning) => {
    set((state) => ({
      sessions: {
        ...state.sessions,
        [sessionId]: touchSession(state.sessions[sessionId], { isRunning }),
      },
    }));
  },
}));
