import { Outlet, createRootRoute, useRouterState } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect, useState } from "react"

import { AppLayout } from "@/components/app-layout"
import { initializeChatStore } from "@/features/chat/controller"
import { isLauncherAuthPathname } from "@/lib/launcher-login-path"

const RootLayout = () => {
  // Prefer the real address bar path: stale embedded bundles may not register
  // /launcher-login or /launcher-setup in the route tree, which would otherwise
  // keep AppLayout + gateway polling → 401 → launcherFetch redirect loop.
  const routerState = useRouterState({
    select: (s) => ({
      pathname: s.location.pathname,
      matches: s.matches,
    }),
  })

  const windowPath =
    typeof globalThis.location !== "undefined"
      ? globalThis.location.pathname || "/"
      : routerState.pathname

  const isAuthPage =
    isLauncherAuthPathname(windowPath) ||
    isLauncherAuthPathname(routerState.pathname) ||
    routerState.matches.some(
      (m) => m.routeId === "/launcher-login" || m.routeId === "/launcher-setup",
    )

  const [authError, setAuthError] = useState<string | null>(null)

  // Auth check disabled - always load the app directly

  useEffect(() => {
    if (isAuthPage) {
      return
    }
    initializeChatStore()
  }, [isAuthPage])

  if (isAuthPage) {
    return (
      <>
        <Outlet />
        {import.meta.env.DEV ? <TanStackRouterDevtools /> : null}
      </>
    )
  }

  return (
    <>
      {authError && (
        <div className="bg-destructive text-destructive-foreground fixed inset-x-0 top-0 z-[100] flex items-center justify-between px-4 py-2 text-sm shadow-md">
          <span>Auth service error: {authError}</span>
          <button
            className="ml-4 opacity-70 hover:opacity-100"
            onClick={() => setAuthError(null)}
            aria-label="Dismiss"
          >
            ✕
          </button>
        </div>
      )}
      <AppLayout>
        <Outlet />
        {import.meta.env.DEV ? <TanStackRouterDevtools /> : null}
      </AppLayout>
    </>
  )
}

export const Route = createRootRoute({ component: RootLayout })
