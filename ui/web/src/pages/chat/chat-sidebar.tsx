import { memo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SessionSwitcher } from "@/components/chat/session-switcher";
import { AgentPickerModal } from "@/components/chat/agent-picker-modal";
import type { SessionInfo } from "@/types/session";

interface ChatSidebarProps {
  sessions: SessionInfo[];
  sessionsLoading: boolean;
  activeSessionKey: string;
  onSessionSelect: (key: string) => void;
  onDeleteSession?: (key: string) => void;
  onNewChatWithAgent: (agentId: string) => void;
}

export const ChatSidebar = memo(function ChatSidebar({
  sessions,
  sessionsLoading,
  activeSessionKey,
  onSessionSelect,
  onDeleteSession,
  onNewChatWithAgent,
}: ChatSidebarProps) {
  const { t } = useTranslation("chat");
  const [agentPickerOpen, setAgentPickerOpen] = useState(false);

  const handleAgentSelected = (agentId: string) => {
    setAgentPickerOpen(false);
    onNewChatWithAgent(agentId);
  };

  return (
    <div className="flex h-full w-72 max-w-[85vw] flex-col border-r bg-background">
      {/* New chat button */}
      <div className="border-b p-3">
        <Button
          variant="outline"
          className="w-full justify-start gap-2"
          onClick={() => setAgentPickerOpen(true)}
        >
          <Plus className="h-4 w-4" />
          {t("newChat")}
        </Button>
      </div>

      {/* Session list */}
      <div className="flex-1 overflow-y-auto">
        <SessionSwitcher
          sessions={sessions}
          activeKey={activeSessionKey}
          onSelect={onSessionSelect}
          onDelete={onDeleteSession}
          loading={sessionsLoading}
        />
      </div>

      {/* Agent picker modal */}
      <AgentPickerModal
        open={agentPickerOpen}
        onOpenChange={setAgentPickerOpen}
        onAgentSelected={handleAgentSelected}
      />
    </div>
  );
});
