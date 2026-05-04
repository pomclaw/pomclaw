import { Suspense } from "react";
import { Routes, Route, Navigate } from "react-router";
import { AppLayout } from "@/components/layout/app-layout";
import { RequireAuth } from "@/components/shared/require-auth";
import { RequireSetup } from "@/components/shared/require-setup";
import { ErrorBoundary } from "@/components/shared/error-boundary";
import { ROUTES } from "@/lib/constants";
import { lazyWithRetry } from "@/lib/lazy-with-retry";

// Lazy-loaded pages
const LoginPage = lazyWithRetry(() =>
  import("@/pages/login/login-page").then((m) => ({ default: m.LoginPage })),
);
const OverviewPage = lazyWithRetry(() =>
  import("@/pages/overview/overview-page").then((m) => ({ default: m.OverviewPage })),
);
const ChatPage = lazyWithRetry(() =>
  import("@/pages/chat/chat-page").then((m) => ({ default: m.ChatPage })),
);
const AgentsPage = lazyWithRetry(() =>
  import("@/pages/agents/agents-page").then((m) => ({ default: m.AgentsPage })),
);
const TeamsPage = lazyWithRetry(() =>
  import("@/pages/teams/teams-page").then((m) => ({ default: m.TeamsPage })),
);
const SetupPage = lazyWithRetry(() =>
  import("@/pages/setup/setup-page").then((m) => ({ default: m.SetupPage })),
);
const TenantSelectorPage = lazyWithRetry(() =>
  import("@/pages/login/tenant-selector").then((m) => ({ default: m.TenantSelectorPage })),
);

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center">
      <img src="/pomclaw-logo.svg" alt="" className="h-8 w-8 animate-pulse opacity-50" />
    </div>
  );
}

export function AppRoutes() {
  return (
    <ErrorBoundary>
    <Suspense fallback={<PageLoader />}>
      <Routes>
        <Route path={ROUTES.LOGIN} element={<LoginPage />} />

        {/* Tenant selector — accessible when authenticated but tenant not yet selected */}
        <Route path={ROUTES.SELECT_TENANT} element={<TenantSelectorPage />} />

        {/* Setup wizard — standalone layout, requires auth but no sidebar */}
        <Route
          path={ROUTES.SETUP}
          element={
            <RequireAuth>
              <SetupPage />
            </RequireAuth>
          }
        />

        {/* Main app — requires auth + setup complete */}
        <Route
          element={
            <RequireAuth>
              <RequireSetup>
                <AppLayout />
              </RequireSetup>
            </RequireAuth>
          }
        >
          <Route index element={<Navigate to={ROUTES.OVERVIEW} replace />} />

          {/* 核心功能：概览 */}
          <Route path={ROUTES.OVERVIEW} element={<OverviewPage />} />

          {/* 核心功能：对话 */}
          <Route path={ROUTES.CHAT_PATTERN} element={<ChatPage />} />

          {/* 核心功能：Agents */}
          <Route path={ROUTES.AGENTS} element={<AgentsPage key="list" />} />
          <Route path={ROUTES.AGENT_DETAIL} element={<AgentsPage key="detail" />} />

          {/* 核心功能：Agent Teams & Links */}
          <Route path={ROUTES.TEAMS} element={<TeamsPage key="list" />} />
          <Route path={ROUTES.TEAM_DETAIL} element={<TeamsPage key="detail" />} />
        </Route>

        {/* Catch-all → overview */}
        <Route path="*" element={<Navigate to={ROUTES.OVERVIEW} replace />} />
      </Routes>
    </Suspense>
    </ErrorBoundary>
  );
}
