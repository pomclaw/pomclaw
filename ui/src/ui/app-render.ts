import { html } from "lit";
import type { AppViewState } from "./app-view-state.ts";
import { renderLoginGate } from "./views/login-gate.ts";
import { renderChat, type ChatProps } from "./views/chat.ts";

/**
 * Main app render function - simplified to only handle login and chat pages.
 * Protocol-compatible with OpenClaw for future migration.
 */
export function renderApp(state: AppViewState) {
  // Show login page until connected
  if (!state.connected) {
    return html`${renderLoginGate(state)}`;
  }

  // Convert AppViewState to ChatProps and render
  const chatProps = appStateToChatProps(state);
  return renderChat(chatProps);
}

/**
 * Convert AppViewState to ChatProps with sensible defaults
 */
function appStateToChatProps(state: AppViewState): ChatProps {
  return {
    sessionKey: state.sessionKey,
    onSessionKeyChange: (next: string) => {
      state.applySettings({ ...state.settings, sessionKey: next });
    },
    thinkingLevel: state.chatThinkingLevel,
    showThinking: state.settings.chatShowThinking ?? false,
    showToolCalls: state.settings.chatShowToolCalls ?? true,
    loading: state.chatLoading,
    sending: state.chatSending,
    canAbort: false,
    messages: state.chatMessages ?? [],
    toolMessages: state.chatToolMessages ?? [],
    streamSegments: state.chatStreamSegments ?? [],
    stream: state.chatStream ?? null,
    streamStartedAt: state.chatStreamStartedAt ?? null,
    assistantAvatarUrl: state.chatAvatarUrl ?? null,
    draft: state.chatMessage ?? "",
    queue: state.chatQueue ?? [],
    connected: state.connected,
    canSend: state.connected && !state.chatLoading && !state.chatSending,
    disabledReason: state.connected ? null : "未连接到网关",
    error: state.lastError ?? null,
    sessions: state.sessionsResult ?? null,
    focusMode: false,
    assistantName: state.assistantName || "Assistant",
    assistantAvatar: state.assistantAvatar ?? null,
    attachments: state.chatAttachments ?? [],
    onAttachmentsChange: (attachments) => {
      // Handle attachments if needed
    },
    onRefresh: async () => {
      await state.handleSessionsLoad?.();
    },
    onToggleFocusMode: () => {
      // Handle focus mode if needed
    },
    onDraftChange: (next: string) => {
      state.setChatMessage(next);
    },
    onSend: () => {
      state.handleSendChat?.();
    },
    onAbort: () => {
      state.handleAbortChat?.();
    },
    onQueueRemove: (id: string) => {
      state.removeQueuedMessage?.(id);
    },
    onNewSession: () => {
      // Handle new session
    },
    onClearHistory: () => {
      // Handle clear history
    },
    agentsList: state.agentsList ?? null,
    currentAgentId: state.agentsList?.defaultId ?? "main",
    onAgentChange: (agentId: string) => {
      // Handle agent change
    },
    onSessionSelect: (sessionKey: string) => {
      state.applySettings({ ...state.settings, sessionKey });
    },
  };
}
