import { useEffect } from "react";
import { Activity, Bot, DollarSign, Hash, Radio, AlertTriangle } from "lucide-react";
import { Link } from "react-router";
import { useTranslation } from "react-i18next";
import { PageHeader } from "@/components/shared/page-header";
import { StatusBadge } from "@/components/shared/status-badge";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { useAuthStore } from "@/stores/use-auth-store";
import { useWsCall } from "@/hooks/use-ws-call";
import { useProviders } from "@/pages/providers/hooks/use-providers";
// import { useTraces } from "@/pages/traces/hooks/use-traces";
import { Methods } from "@/api/protocol";
import { ROUTES } from "@/lib/constants";
import { formatTokens, formatCost } from "@/lib/format";
import { useSystemHealth } from "@/api/system-health";

import type {
  HealthPayload,
  QuotaUsageResult,
  ChannelStatusPayload,
} from "./types";
import { useLiveUptime } from "./hooks/use-live-uptime";
import { StatCard } from "./stat-card";
import { useOverviewSparklines } from "./hooks/use-overview-sparklines";
import { SystemHealthCard } from "./system-health-card";
import { RecentRequestsCard } from "./recent-requests-card";
import { QuotaUsageCard } from "./quota-usage-card";
// import { useRuntimes } from "@/pages/skills/hooks/use-runtimes";
// import {
//   getChannelAttentionPriority,
//   getChannelStatusFallback,
// } from "@/pages/channels/channels-status-view";
// import { useChannelInstances } from "@/pages/channels/hooks/use-channel-instances";

// const UsagePage = lazy(() =>
//   import("@/pages/usage/usage-page").then((m) => ({ default: m.UsagePage })),
// );

const REFRESH_INTERVAL = 30_000;
// const MAX_OVERVIEW_CHANNEL_INSTANCES = 200;

export function OverviewPage() {
  const { t } = useTranslation("overview");
  const connected = useAuthStore((s) => s.connected);
  const { data: health } = useSystemHealth();
  const { call: fetchQuota, data: quota } =
    useWsCall<QuotaUsageResult>(Methods.QUOTA_USAGE);
  const sparklines = useOverviewSparklines();
  const { call: fetchChannels, data: channelStatusData } =
    useWsCall<ChannelStatusPayload>(Methods.CHANNELS_STATUS);
  const { providers, loading: providersLoading } = useProviders();
  // const { runtimes } = useRuntimes();
  const runtimes = { runtimes: [] };
  // const { traces } = useTraces({ limit: 8 });
  const traces: any[] = [];
  // const { instances: channelInstances, total: channelInstanceTotal } = useChannelInstances({
  //   limit: MAX_OVERVIEW_CHANNEL_INSTANCES,
  //   offset: 0,
  // });
  // const channelInstances: any[] = [];
  const channelInstanceTotal = 0;

  // Use backend health data to check if providers are configured
  const hasNoProviders = (health?.providers ?? 0) === 0;
  const hasNoEnabledProviders = false; // Not used with new backend data

  useEffect(() => {
    if (!connected) return;
    fetchQuota();
    fetchChannels();
    const id = setInterval(() => {
      fetchQuota();
      fetchChannels();
    }, REFRESH_INTERVAL);
    return () => clearInterval(id);
  }, [connected, fetchQuota, fetchChannels]);

  const liveUptime = useLiveUptime(health?.uptime);

  // Computed
  const agentTotal = 0; // Not available from /v1/system/health, will be enhanced later
  const runningAgents = 0;
  const channelStatusMap = channelStatusData?.channels ?? {};
  // const canSynthesizeOverviewFallbacks =
  //   channelInstanceTotal > 0 &&
  //   channelInstanceTotal <= MAX_OVERVIEW_CHANNEL_INSTANCES &&
  //   channelInstances.length >= channelInstanceTotal;
  const channelEntries = Object.entries(channelStatusMap);
  // const channelEntries = (() => {
  //   const combined = new Map(Object.entries(channelStatusMap));
  //   if (canSynthesizeOverviewFallbacks) {
  //     for (const instance of channelInstances) {
  //       if (combined.has(instance.name)) continue;
  //       const fallback = getChannelStatusFallback(instance);
  //       if (fallback) {
  //         combined.set(instance.name, fallback);
  //       }
  //     }
  //   }
  //   return [...combined.entries()];
  // })();
  const totalChannelCount = Math.max(channelEntries.length, channelInstanceTotal);
  const channelsOnline = channelEntries.filter(([, c]) => c.running).length;
  // const channelsNeedingAttention = channelEntries.filter(
  //   ([, c]) => getChannelAttentionPriority(c, c.enabled) > 0,
  // ).length;
  // const overviewAttentionCount = canSynthesizeOverviewFallbacks
  //   ? channelsNeedingAttention
  //   : null;
  const overviewAttentionCount = null;
  const enabledProviders = providers.filter((p) => p.enabled);

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
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
            <StatCard
              icon={Activity}
              label={t("statCards.requestsToday")}
              value={quota?.requestsToday ?? 0}
              sub={
                quota?.uniqueUsersToday
                  ? t("statCards.users", { count: quota.uniqueUsersToday })
                  : undefined
              }
              sparkline={sparklines?.requestSparkline}
              trend={sparklines?.trends.requests}
            />
            <StatCard
              icon={Hash}
              label={t("statCards.tokensToday")}
              value={formatTokens(
                (quota?.inputTokensToday ?? 0) + (quota?.outputTokensToday ?? 0),
              )}
              sub={
                quota
                  ? t("statCards.inOut", { input: formatTokens(quota.inputTokensToday), output: formatTokens(quota.outputTokensToday) })
                  : undefined
              }
              sparkline={sparklines?.tokenSparkline}
              trend={sparklines?.trends.tokens}
            />
            <StatCard
              icon={DollarSign}
              label={t("statCards.costToday", "Cost Today")}
              value={formatCost(quota?.costToday)}
              sparkline={sparklines?.costSparkline}
              trend={sparklines?.trends.cost}
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
            <StatCard
              icon={Radio}
              label={t("statCards.channels")}
              value={
                totalChannelCount > 0
                  ? `${channelsOnline} / ${totalChannelCount}`
                  : "0"
              }
              sub={
                totalChannelCount > 0
                  ? overviewAttentionCount && overviewAttentionCount > 0
                    ? t("statCards.channelsAttention", {
                        defaultValue: "{{count}} need attention",
                        count: overviewAttentionCount,
                      })
                    : t("statCards.online")
                  : undefined
              }
            />
          </div>

          {/* System Health */}
          <SystemHealthCard
            health={health as HealthPayload | null}
            liveUptime={liveUptime}
            enabledProviderCount={health?.providers ?? enabledProviders.length}
            sessions={health?.sessions ?? 0}
            channelEntries={channelEntries}
            runtimeEntries={runtimes?.runtimes}
          />

          {/* Recent Requests */}
          <RecentRequestsCard traces={traces} />

          {/* Quota Usage */}
          {quota?.enabled && quota.entries.length > 0 && (
            <QuotaUsageCard quota={quota} />
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
