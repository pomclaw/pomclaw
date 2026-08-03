import { useTranslation } from "react-i18next";
import type { ToolPolicyConfig } from "@/types/agent";
import { ToolPolicySection } from "../config-sections";
import { ConfigGroupHeader } from "@/components/shared/config-group-header";

interface CapabilitiesSectionProps {
  toolsEnabled: boolean;
  tools: ToolPolicyConfig;
  onToolsToggle: (v: boolean) => void;
  onToolsChange: (v: ToolPolicyConfig) => void;
}

export function CapabilitiesSection({
  toolsEnabled, tools, onToolsToggle, onToolsChange,
}: CapabilitiesSectionProps) {
  const { t } = useTranslation("agents");

  return (
    <section className="space-y-4 rounded-lg border p-3 sm:p-4">
      <ConfigGroupHeader
        title={t("detail.capabilities")}
        description={t("configGroups.capabilitiesDesc")}
      />
      <div className="space-y-4">
        <ToolPolicySection
          enabled={toolsEnabled}
          value={tools}
          onToggle={(v) => { onToolsToggle(v); if (!v) onToolsChange({}); }}
          onChange={onToolsChange}
        />
      </div>
    </section>
  );
}
