import { Activity, Bot, DollarSign, Hash, AlertTriangle } from "lucide-react";
import { Link } from "react-router";
import { useTranslation } from "react-i18next";
import { PageHeader } from "@/components/shared/page-header";
import { StatusBadge } from "@/components/shared/status-badge";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { useAuthStore } from "@/stores/use-auth-store";
import { ROUTES } from "@/lib/constants";
import { formatTokens, formatCost } from "@/lib/format";
import { useSystemHealth } from "@/hooks/use-system-health";

import type { HealthPayload } from "./types";
import { useLiveUptime } from "./hooks/use-live-uptime";
import { StatCard } from "./stat-card";
import { SystemHealthCard } from "./system-health-card";
import { RecentRequestsCard } from "./recent-requests-card";
import { useTraces } from "@/pages/traces/hooks/use-traces";
// import {
//   getChannelAttentionPriority,
//   getChannelStatusFallback,
// } from "@/pages/channels/channels-status-view";
// import { useChannelInstances } from "@/pages/channels/hooks/use-channel-instances";

// const UsagePage = lazy(() =>
//   import("@/pages/usage/usage-page").then((m) => ({ default: m.UsagePage })),
// );

export function OverviewPage() {
  const { t } = useTranslation("overview");
  const connected = useAuthStore((s) => s.connected);
  const { data } = useSystemHealth();
  const health = data?.health;
  const quotaUsage = data?.quotaUsage;
  const { traces } = useTraces({ limit: 8 });

  // Use backend health data to check if providers are configured
  const hasNoProviders = (health?.providers ?? 0) === 0;
  const hasNoEnabledProviders = false; // Not used with new backend data

  const liveUptime = useLiveUptime(health?.uptime);

  // Computed
  const agentTotal = health?.agents ?? 0;
  const runningAgents = agentTotal; // All agents shown as running for now

  return (
    <div className="space-y-6 p-4 sm:p-6">
      {/* Header */}
      <PageHeader
        title={t("title")}
        description={t("description")}
        actions={
          <div className="flex items-center gap-2">
            {health?.version && (
              <span className="text-xs text-muted-foreground">
                {health.version}
              </span>
            )}
            <StatusBadge
              status={connected ? "success" : "error"}
              label={connected ? t("common:connected", "Connected") : t("common:disconnected", "Disconnected")}
            />
          </div>
        }
      />

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">{t("tabs.overview")}</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {/* Provider warning */}
          {(hasNoProviders || hasNoEnabledProviders) && (
            <Alert>
              <AlertTriangle className="h-4 w-4" />
              <AlertTitle>
                {hasNoProviders
                  ? t("providers.noProvidersTitle")
                  : t("providers.noEnabledTitle")}
              </AlertTitle>
              <AlertDescription>
                {hasNoProviders
                  ? t("providers.noProvidersDesc")
                  : t("providers.noEnabledDesc")}
                <Link
                  to={ROUTES.PROVIDERS}
                  className="font-medium underline underline-offset-4 hover:text-foreground"
                >
                  {t("providers.goToSettings")}
                </Link>
              </AlertDescription>
            </Alert>
          )}

          {/* Summary cards */}
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              icon={Activity}
              label={t("statCards.requestsToday")}
              value={quotaUsage?.requests_today ?? 0}
            />
            <StatCard
              icon={Hash}
              label={t("statCards.tokensToday")}
              value={formatTokens(
                (quotaUsage?.input_tokens_today ?? 0) + (quotaUsage?.output_tokens_today ?? 0),
              )}
            />
            <StatCard
              icon={DollarSign}
              label={t("statCards.costToday", "Cost Today")}
              value={formatCost(quotaUsage?.cost_today)}
            />
            <StatCard
              icon={Bot}
              label={t("statCards.agents")}
              value={
                agentTotal > 0
                  ? `${runningAgents} / ${agentTotal}`
                  : "0"
              }
              sub={agentTotal > 0 ? t("statCards.running") : undefined}
            />
          </div>

          {/* System Health */}
          <SystemHealthCard
            health={health as HealthPayload | null}
            liveUptime={liveUptime}
            enabledProviderCount={health?.providers ?? 0}
            sessions={health?.sessions ?? 0}
            channelEntries={[]}
          />

          {/* Recent Requests */}
          <RecentRequestsCard traces={traces} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
