import { Suspense } from "react";
import { Routes, Route, Navigate } from "react-router";
import { ROUTES } from "@/lib/constants";
import { lazyWithRetry } from "@/lib/lazy-with-retry";

// Lazy-loaded pages
const TracesPage = lazyWithRetry(() =>
  import("@/pages/traces/traces-page").then((m) => ({ default: m.TracesPage })),
);

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="h-8 w-8 animate-pulse opacity-50">Loading...</div>
    </div>
  );
}

export function AppRoutes() {
  return (
    <Suspense fallback={<PageLoader />}>
      <Routes>
        <Route index element={<Navigate to={ROUTES.OVERVIEW} replace />} />
        <Route path={ROUTES.TRACES} element={<TracesPage key="list" />} />
        <Route path={ROUTES.TRACE_DETAIL} element={<TracesPage key="detail" />} />
        {/* Catch-all → overview */}
        <Route path="*" element={<Navigate to={ROUTES.OVERVIEW} replace />} />
      </Routes>
    </Suspense>
  );
}
