import { useState } from "react";
import { useNavigate } from "react-router";
import { useTranslation } from "react-i18next";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAgentDetail } from "../hooks/use-agent-detail";
import { useAgents } from "../hooks/use-agents";
import { useAgentHeartbeat } from "../hooks/use-agent-heartbeat";
import { AgentHeader } from "./agent-header";
import { AgentOverviewTab } from "./agent-overview-tab";
import { AgentFilesTab } from "./agent-files-tab";
import { AgentSkillsTab } from "./agent-skills-tab";
import { SummoningModal } from "../summoning-modal";
import { ConfirmDeleteDialog } from "@/components/shared/confirm-delete-dialog";
import { DetailPageSkeleton } from "@/components/shared/loading-skeleton";
import { agentDisplayName } from "./agent-display-utils";
import { SystemPromptDialog } from "./system-prompt-dialog";

interface AgentDetailPageProps {
  agentId: string;
  onBack: () => void;
}

export function AgentDetailPage({ agentId, onBack }: AgentDetailPageProps) {
  const { t } = useTranslation("agents");
  const navigate = useNavigate();
  const { agent, files, loading, updateAgent, getFile, setFile, resummonAgent, refresh } =
    useAgentDetail(agentId);
  const { deleteAgent: deleteAgentById, cancelSummonAgent } = useAgents();
  const hb = useAgentHeartbeat(agentId);
  const [summoningOpen, setSummoningOpen] = useState(false);
  const [activeTab, setActiveTab] = useState("agent");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [promptOpen, setPromptOpen] = useState(false);

  const handleSummoningClose = (open: boolean) => {
    setSummoningOpen(open);
    if (!open) refresh();
  };

  if (loading || !agent) {
    return <DetailPageSkeleton tabs={3} />;
  }

  const title = agentDisplayName(agent, t("card.unnamedAgent"));

  return (
    <div>
      <AgentHeader
        agent={agent}
        heartbeat={hb.config}
        onBack={onBack}
        onDelete={() => setDeleteOpen(true)}
        onHeartbeat={() => {}}
        onSystemPrompt={() => setPromptOpen(true)}
      />

      <div className="p-3 sm:p-4">
        <div className="max-w-4xl">
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="w-full justify-start overflow-x-auto overflow-y-hidden">
              <TabsTrigger value="agent">{t("detail.tabs.agent")}</TabsTrigger>
              <TabsTrigger value="files">{t("detail.tabs.files")}</TabsTrigger>
              <TabsTrigger value="skills">{t("detail.tabs.skills")}</TabsTrigger>
            </TabsList>

            <TabsContent value="agent" className="mt-4">
              <AgentOverviewTab
                key={agent.id + "-" + agent.updated_at}
                agent={agent}
                onUpdate={updateAgent}
                heartbeat={hb}
                onManageCodexPool={() => navigate(`/agents/${agent.id}/codex-pool`)}
              />
            </TabsContent>

            <TabsContent value="files" className="mt-4">
              <AgentFilesTab
                agent={agent}
                files={files}
                onGetFile={getFile}
                onSetFile={setFile}
              />
            </TabsContent>

            <TabsContent value="skills" className="mt-4">
              <AgentSkillsTab
                agentId={agentId}
                agent={agent}
                onUpdate={updateAgent}
              />
            </TabsContent>
          </Tabs>
        </div>
      </div>

      <SummoningModal
        open={summoningOpen}
        onOpenChange={handleSummoningClose}
        agentId={agentId}
        agentName={title}
        onCompleted={() => {}}
        onResummon={async () => { await resummonAgent(); }}
        onCancel={cancelSummonAgent}
      />

      {promptOpen && (
        <SystemPromptDialog
          agentKey={agentId}
          open={promptOpen}
          onOpenChange={setPromptOpen}
        />
      )}

      <ConfirmDeleteDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={t("delete.title")}
        description={t("delete.detailDescription", { name: title })}
        confirmValue={agent.display_name || agent.agent_key}
        confirmLabel={t("delete.confirmLabel")}
        onConfirm={async () => {
          await deleteAgentById(agentId);
          setDeleteOpen(false);
          onBack();
        }}
      />
    </div>
  );
}