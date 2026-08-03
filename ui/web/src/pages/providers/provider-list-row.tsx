import { Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ChatGPTOAuthQuotaStrip } from "@/pages/agents/agent-detail/chatgpt-oauth-quota-strip";
import type { EffectiveChatGPTOAuthRoutingStrategy } from "@/types/agent";
import { getProviderReasoningDefaults } from "@/types/provider";
import type { ChatGPTOAuthProviderQuota } from "./hooks/use-chatgpt-oauth-provider-quotas";
import type { ChatGPTOAuthAvailability } from "./hooks/use-chatgpt-oauth-provider-statuses";
import type { Provider } from "./hooks/use-providers";
import { PROVIDER_TYPE_BADGE, ProviderApiKeyBadge } from "./provider-utils";

interface ProviderOAuthPoolSummary {
  availability: ChatGPTOAuthAvailability;
  role: "owner" | "member" | "standalone";
  managedByLabel?: string;
  memberCount: number;
  strategy: EffectiveChatGPTOAuthRoutingStrategy;
  quota?: ChatGPTOAuthProviderQuota | null;
  quotaLoading?: boolean;
}

interface ProviderListRowProps {
  provider: Provider;
  oauthPool?: ProviderOAuthPoolSummary;
  showPoolHint?: boolean;
  onClick: () => void;
  onDelete?: () => void;
  onPoolSetup?: () => void;
}

function strategyLabelKey(strategy: EffectiveChatGPTOAuthRoutingStrategy): string {
  if (strategy === "round_robin") return "list.strategy.roundRobin";
  return "list.strategy.priorityOrder";
}

export function ProviderListRow({
  provider,
  oauthPool,
  showPoolHint,
  onClick,
  onDelete,
  onPoolSetup,
}: ProviderListRowProps) {
  const { t: tc } = useTranslation("common");
  const { t } = useTranslation("providers");
  const typeBadge = PROVIDER_TYPE_BADGE[provider.provider_type] ?? {
    label: provider.provider_type,
    variant: "outline" as const,
  };
  const showAvailabilityWarning = oauthPool && oauthPool.availability !== "ready";
  const hasPoolRole = oauthPool?.role === "owner" || oauthPool?.role === "member";
  const reasoningDefaults = getProviderReasoningDefaults(provider.settings);
  const showQuota = provider.provider_type === "chatgpt_oauth"
    && (oauthPool?.quotaLoading || Boolean(oauthPool?.quota));

  return (
    <tr
      onClick={onClick}
      className={cn(
        "cursor-pointer border-b last:border-0 hover:bg-muted/30",
        oauthPool?.role === "owner" && "bg-primary/[0.02]",
        oauthPool?.role === "member" && "bg-sky-500/[0.025]",
      )}
    >
      {/* Name column */}
      <td className="px-3 py-2.5">
        <div className="flex items-center gap-2">
          <span className="font-medium">{provider.name}</span>
          <span
            className={cn(
              "inline-block h-2 w-2 shrink-0 rounded-full",
              provider.enabled ? "bg-emerald-500" : "bg-muted-foreground/40",
            )}
          />
          {provider.is_shared && (
            <Badge variant="secondary" className="h-5 px-1.5 text-2xs">
              {tc("shared")}
            </Badge>
          )}
          {hasPoolRole && (
            <Badge
              variant={oauthPool.role === "owner" ? "outline" : "info"}
              className={cn(
                "h-5 px-1.5 text-2xs",
                oauthPool.role === "owner" && "border-primary/30 bg-primary/[0.06] text-primary",
              )}
            >
              {t(oauthPool.role === "owner" ? "list.poolOwner" : "list.poolMember")}
            </Badge>
          )}
          {showPoolHint && !hasPoolRole && onPoolSetup ? (
            <Badge
              variant="outline"
              className="h-5 cursor-pointer border-dashed border-primary/40 px-1.5 text-2xs text-primary transition-colors hover:border-primary hover:bg-primary/10"
              onClick={(event) => { event.stopPropagation(); onPoolSetup(); }}
            >
              {t("list.poolAvailable")}
            </Badge>
          ) : null}
          {reasoningDefaults ? (
            <Badge variant="secondary" className="h-5 px-1.5 text-2xs">
              {t("list.reasoningDefault", {
                level: t(`reasoning.${reasoningDefaults.effort ?? "off"}`),
              })}
            </Badge>
          ) : null}
        </div>
        {provider.description && (
          <div className="mt-0.5 text-xs text-muted-foreground">{provider.description}</div>
        )}
        {(showAvailabilityWarning || showQuota) && (
          <div className="mt-0.5 flex items-center gap-2 text-xs">
            {showAvailabilityWarning && (
              <span
                className={cn(
                  "shrink-0 font-medium",
                  oauthPool?.availability === "disabled"
                    ? "text-muted-foreground"
                    : "text-amber-700 dark:text-amber-400",
                )}
              >
                {t(
                  oauthPool?.availability === "disabled"
                    ? "list.status.disabled"
                    : "list.status.needsSignIn",
                )}
              </span>
            )}
            {showQuota && (
              <ChatGPTOAuthQuotaStrip
                quota={oauthPool?.quota}
                loading={oauthPool?.quotaLoading}
                compact
                layout="inline"
                embedded
                translationNamespace="providers"
                translationKeyPrefix="quota"
                className="shrink-0"
              />
            )}
          </div>
        )}
      </td>

      {/* Type column */}
      <td className="px-3 py-2.5">
        <Badge variant={typeBadge.variant} className="text-xs-plus">
          {typeBadge.label}
        </Badge>
      </td>

      {/* API Key column */}
      <td className="px-3 py-2.5">
        <ProviderApiKeyBadge provider={provider} oauthAvailability={oauthPool?.availability} />
      </td>

      {/* Status column */}
      <td className="px-3 py-2.5">
        <Badge variant={provider.enabled ? "default" : "secondary"} className="text-xs">
          {provider.enabled ? tc("enabled") : tc("disabled")}
        </Badge>
      </td>

      {/* Created by column */}
      <td className="px-3 py-2.5 text-muted-foreground">
        {provider.created_by_name || provider.created_by || "-"}
      </td>

      {/* Actions column */}
      <td className="px-3 py-2.5">
        <div className="flex items-center justify-end gap-1">
          {onDelete && (
            <Button
              variant="ghost"
              size="icon-sm"
              className="text-muted-foreground hover:text-destructive"
              onClick={(event) => {
                event.stopPropagation();
                onDelete();
              }}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      </td>
    </tr>
  );
}