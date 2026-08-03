import { useState, useCallback, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useParams, useNavigate } from "react-router";
import { PanelLeftOpen } from "lucide-react";
import { useAuthStore } from "@/stores/use-auth-store";
import { useIsMobile } from "@/hooks/use-media-query";
import { cn } from "@/lib/utils";
import { ChatSidebar } from "./chat-sidebar";
import { ChatThread } from "./chat-thread";
import { ChatInput, type AttachedFile } from "@/components/chat/chat-input";
import { ChatTopBar } from "@/components/chat/chat-top-bar";
import { DropZone } from "@/components/chat/drop-zone";
import { AgentPickerPrompt } from "@/components/chat/agent-picker-prompt";
import { useApiClient } from "@/hooks/use-api-client";
import { useChatSessions } from "./hooks/use-chat-sessions";
import { useChatMessages } from "./hooks/use-chat-messages";
import { useChatSend } from "./hooks/use-chat-send";
import { useVirtualKeyboard } from "@/hooks/use-virtual-keyboard";
import { TaskPanel } from "@/components/chat/task-panel";

export function ChatPage() {
  const { t } = useTranslation("chat");
  const { sessionId: urlSessionKey } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const connected = useAuthStore((s) => s.connected);
  const api = useApiClient();

  const [scrollTrigger, setScrollTrigger] = useState(0);
  const [files, setFiles] = useState<AttachedFile[]>([]);

  // sessionId derived from URL — single source of truth, no separate state
  const sessionId = urlSessionKey ?? "";

  // Fallback agent ID used only when URL has no session key
  const [agentIdFallback, setAgentIdFallback] = useState("");

  // Start with fallback, will be updated from sessions if available
  const [agentId, setAgentId] = useState(agentIdFallback);

  // Agent is confirmed when URL has a session (agentId parsed) or user explicitly picked one
  const agentConfirmed = !!urlSessionKey || !!agentIdFallback;

  const {
    sessions,
    loading: sessionsLoading,
    refresh: refreshSessions,
    buildNewSessionKey,
    deleteSession,
  } = useChatSessions();

  // Update agentId from sessions or fallback
  useEffect(() => {
    if (urlSessionKey && !urlSessionKey.startsWith("new:")) {
      // Try to find agentId from loaded sessions
      const session = sessions.find((s) => s.key === urlSessionKey);
      if (session && session.agentId) {
        setAgentId(session.agentId);
        return;
      }
    }
    // Use fallback if no session found or new session
    if (agentIdFallback) {
      setAgentId(agentIdFallback);
    }
  }, [urlSessionKey, sessions, agentIdFallback]);

  const {
    messages,
    streamText,
    thinkingText,
    toolStream,
    isRunning,
    isBusy,
    loading: messagesLoading,
    activity,
    blockReplies,
    teamTasks,
    expectRun,
    addLocalMessage,
  } = useChatMessages(sessionId, agentId);

  // Refresh sessions when all work completes (main agent + team tasks)
  const prevIsBusyRef = useRef(false);
  useEffect(() => {
    if (prevIsBusyRef.current && !isBusy) {
      refreshSessions();
    }
    prevIsBusyRef.current = isBusy;
  }, [isBusy, refreshSessions]);


  const handleMessageAdded = useCallback(
    (msg: { role: "user" | "assistant" | "tool"; content: string; timestamp?: number }, key?: string) => {
      addLocalMessage(msg, key);
    },
    [addLocalMessage],
  );

  const { send, abort, error: sendError } = useChatSend({
    agentId,
    onMessageAdded: handleMessageAdded,
    onExpectRun: expectRun,
  });

  const handleNewChatWithAgent = useCallback(
    (selectedAgentId: string) => {
      setAgentIdFallback(selectedAgentId);
      // 创建一个新 session 的占位符
      navigate(`/chat/${encodeURIComponent(buildNewSessionKey())}`);
    },
    [buildNewSessionKey, navigate],
  );

  const handleSessionSelect = useCallback(
    (key: string) => {
      navigate(`/chat/${encodeURIComponent(key)}`);
    },
    [navigate],
  );

  const handleDeleteSession = useCallback(async (key: string) => {
    await deleteSession(key);
    if (key === sessionId) {
      const next = sessions.find((s) => s.key !== key);
      if (next) {
        handleSessionSelect(next.key);
      } else {
        // 如果删除了当前 session，返回对话列表
        navigate("/chat");
      }
    }
  }, [deleteSession, sessionId, sessions, handleSessionSelect, navigate]);

  const handleSend = useCallback(
    async (message: string, sendFiles?: AttachedFile[]) => {
      if (!agentId) return;

      let key = sessionId;

      // If this is a new (unsaved) session, create it via HTTP first to get a real numeric ID.
      // Backend's chat.send expects sessionId as int64, not the "new:uuid" placeholder.
      if (!key || key.startsWith("new:")) {
        try {
          const res = await api.createSession({ agent_id: agentId });
          if (!res.session?.id) return;
          key = String(res.session.id);
          navigate(`/chat/${key}`, { replace: true });
          refreshSessions();
        } catch {
          return;
        }
      }

      send(message, key, sendFiles);
      setScrollTrigger((n) => n + 1);
    },
    [sessionId, agentId, api, send, navigate, refreshSessions],
  );

  const handleDropFiles = useCallback((dropped: File[]) => {
    setFiles((prev) => [...prev, ...dropped.map((f) => ({ file: f }))]);
  }, []);

  const handleAbort = useCallback(() => {
    abort(sessionId);
  }, [abort, sessionId]);

  const isMobile = useIsMobile();
  useVirtualKeyboard();
  const [chatSidebarOpen, setChatSidebarOpen] = useState(false);
  const [taskPanelOpen, setTaskPanelOpen] = useState(false);

  // Auto-open task panel when first task appears, auto-close when all done.
  const prevTaskCountRef = useRef(0);
  useEffect(() => {
    const prev = prevTaskCountRef.current;
    const curr = teamTasks.length;
    if (prev === 0 && curr > 0) setTaskPanelOpen(true);
    if (curr === 0 && prev > 0) setTaskPanelOpen(false);
    prevTaskCountRef.current = curr;
  }, [teamTasks.length]);

  const handleSessionSelectMobile = useCallback(
    (key: string) => {
      handleSessionSelect(key);
      setChatSidebarOpen(false);
    },
    [handleSessionSelect],
  );

  const handleNewChatWithAgentMobile = useCallback(
    (agentId: string) => {
      handleNewChatWithAgent(agentId);
      setChatSidebarOpen(false);
    },
    [handleNewChatWithAgent],
  );

  return (
    <div className="relative flex h-full overflow-hidden">
      {/* Chat Sidebar */}
      {isMobile ? (
        <>
          {chatSidebarOpen && (
            <div
              className="fixed inset-0 z-40 bg-black/50"
              onClick={() => setChatSidebarOpen(false)}
            />
          )}
          <div
            className={cn(
              "fixed inset-y-0 left-0 z-50 transition-transform duration-200 ease-in-out",
              chatSidebarOpen ? "translate-x-0" : "-translate-x-full",
            )}
          >
            <ChatSidebar
              sessions={sessions}
              sessionsLoading={sessionsLoading}
              activeSessionKey={sessionId}
              onSessionSelect={handleSessionSelectMobile}
              onDeleteSession={handleDeleteSession}
              onNewChatWithAgent={handleNewChatWithAgentMobile}
            />
          </div>
        </>
      ) : (
        <ChatSidebar
          sessions={sessions}
          sessionsLoading={sessionsLoading}
          activeSessionKey={sessionId}
          onSessionSelect={handleSessionSelect}
          onDeleteSession={handleDeleteSession}
          onNewChatWithAgent={handleNewChatWithAgent}
        />
      )}

      {/* Main chat area */}
      <div className="flex min-w-0 flex-1 min-h-0 flex-col">
        {isMobile && (
          <div className="flex shrink-0 items-center border-b px-3 py-2 landscape-compact">
            <button
              onClick={() => setChatSidebarOpen(true)}
              className="rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
              title={t("openSessions")}
            >
              <PanelLeftOpen className="h-4 w-4" />
            </button>
          </div>
        )}

        <div className="shrink-0">
          <ChatTopBar
            agentId={agentId}
            isRunning={isRunning}
            isBusy={isBusy}
            activity={activity}
            teamTasks={teamTasks}
            onToggleTaskPanel={() => setTaskPanelOpen((v) => !v)}
            taskPanelOpen={taskPanelOpen}
            session={sessions.find((s) => s.key === sessionId) ?? null}
          />
        </div>

        {sendError && (
          <div className="shrink-0 border-b bg-destructive/10 px-4 py-2 text-sm text-destructive">
            {sendError}
          </div>
        )}

        <DropZone onDrop={handleDropFiles}>
          <ChatThread
            messages={messages}
            streamText={streamText}
            thinkingText={thinkingText}
            toolStream={toolStream}
            blockReplies={blockReplies}
            activity={activity}
            teamTasks={teamTasks}
            isRunning={isRunning}
            isBusy={isBusy}
            loading={messagesLoading}
            scrollTrigger={scrollTrigger}
            onToggleTaskPanel={() => setTaskPanelOpen((v) => !v)}
          />

          {!agentConfirmed ? (
            <AgentPickerPrompt onSelect={handleNewChatWithAgent} />
          ) : (
            <ChatInput
              onSend={handleSend}
              onAbort={handleAbort}
              isBusy={isBusy}
              disabled={!connected || !agentId}
              files={files}
              onFilesChange={setFiles}
            />
          )}
        </DropZone>
      </div>

      {/* Mobile overlay backdrop — must render before TaskPanel so panel sits above */}
      {isMobile && taskPanelOpen && (
        <div className="fixed inset-0 z-40 bg-black/50" onClick={() => setTaskPanelOpen(false)} />
      )}

      {/* Task panel — toggleable sidebar on the right */}
      <TaskPanel tasks={teamTasks} open={taskPanelOpen} onClose={() => setTaskPanelOpen(false)} />
    </div>
  );
}
