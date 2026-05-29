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
const SetupPage = lazyWithRetry(() =>
  import("@/pages/setup/setup-page").then((m) => ({ default: m.SetupPage })),
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
const ProvidersPage = lazyWithRetry(() =>
  import("@/pages/providers/providers-page").then((m) => ({ default: m.ProvidersPage })),
);
const SkillsPage = lazyWithRetry(() =>
  import("@/pages/skills/skills-page").then((m) => ({ default: m.SkillsPage })),
);
const BuiltinToolsPage = lazyWithRetry(() =>
  import("@/pages/builtin-tools/builtin-tools-page").then((m) => ({ default: m.BuiltinToolsPage })),
);
const TracesPage = lazyWithRetry(() =>
  import("@/pages/traces/traces-page").then((m) => ({ default: m.TracesPage })),
);
const MemoryPage = lazyWithRetry(() =>
  import("@/pages/memory/memory-page").then((m) => ({ default: m.MemoryPage })),
);

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center">
      <img src="/pomclaw-icon.svg" alt="" className="h-8 w-8 animate-pulse opacity-50" />
    </div>
  );
}

export function AppRoutes() {
  return (
    <ErrorBoundary>
    <Suspense fallback={<PageLoader />}>
      <Routes>
        {/* Public route - no auth required */}
        <Route path={ROUTES.LOGIN} element={<LoginPage />} />

        {/* Auth-only routes - need login but bypass setup check */}
        <Route
          path={ROUTES.SETUP}
          element={
            <RequireAuth>
              <SetupPage />
            </RequireAuth>
          }
        />

        {/* Main app - full auth + setup required */}
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
          <Route path={ROUTES.OVERVIEW} element={<OverviewPage />} />
          <Route path={ROUTES.CHAT_PATTERN} element={<ChatPage />} />
          <Route path={ROUTES.AGENTS} element={<AgentsPage key="list" />} />
          <Route path={ROUTES.AGENT_DETAIL} element={<AgentsPage key="detail" />} />
          <Route path={ROUTES.PROVIDERS} element={<ProvidersPage key="list" />} />
          <Route path={ROUTES.PROVIDER_DETAIL} element={<ProvidersPage key="detail" />} />
          <Route path={ROUTES.SKILLS} element={<SkillsPage />} />
          <Route path={ROUTES.BUILTIN_TOOLS} element={<BuiltinToolsPage />} />
          <Route path={ROUTES.MEMORY} element={<MemoryPage />} />
          <Route path={ROUTES.TRACES} element={<TracesPage key="list" />} />
          <Route path={ROUTES.TRACE_DETAIL} element={<TracesPage key="detail" />} />
        </Route>

        {/* Catch-all → overview */}
        <Route path="*" element={<Navigate to={ROUTES.OVERVIEW} replace />} />
      </Routes>
    </Suspense>
    </ErrorBoundary>
  );
}
