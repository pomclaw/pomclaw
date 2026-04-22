import { useNavigate, Outlet, createRootRoute, useRouterState } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect, useState } from "react"

import { AppLayout } from "@/components/app-layout"
import { SidebarProvider } from "@/components/ui/sidebar"
import { initializeChatStore } from "@/features/chat/controller"
import { isLauncherAuthPathname } from "@/lib/launcher-login-path"
import { useAuth } from "@/hooks/use-auth"

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

  const isLauncherAuthPage =
    isLauncherAuthPathname(windowPath) ||
    isLauncherAuthPathname(routerState.pathname) ||
    routerState.matches.some(
      (m) => m.routeId === "/launcher-login" || m.routeId === "/launcher-setup",
    )

  const isGatewayAuthPage =
    routerState.pathname === "/login" ||
    routerState.pathname === "/register" ||
    routerState.matches.some(
      (m) => m.routeId === "/login" || m.routeId === "/register",
    )

  const isAuthPage = isLauncherAuthPage || isGatewayAuthPage

  const { isAuthenticated, restoreSession, isLoading } = useAuth()
  const navigate = useNavigate()
  const [authError, setAuthError] = useState<string | null>(null)

  // Restore session on app load
  useEffect(() => {
    restoreSession()
  }, [restoreSession])

  // Initialize chat store if authenticated and not on auth page
  useEffect(() => {
    if (isAuthPage || isLoading) {
      return
    }
    if (isAuthenticated) {
      initializeChatStore()
    }
  }, [isAuthPage, isAuthenticated, isLoading])

  // Redirect to login if not authenticated and not on auth page
  useEffect(() => {
    if (isLoading) {
      return // Wait for session restoration
    }
    if (!isAuthPage && !isAuthenticated) {
      navigate({ to: "/login", replace: true })
    }
  }, [isAuthPage, isAuthenticated, isLoading, navigate])

  // Render content based on auth state
  const renderContent = () => {
    // For auth pages, render without AppLayout
    if (isAuthPage) {
      return (
        <>
          <Outlet />
          {import.meta.env.DEV ? <TanStackRouterDevtools /> : null}
        </>
      )
    }

    // Show loading while restoring session
    if (isLoading) {
      return (
        <div className="bg-background flex min-h-dvh items-center justify-center">
          <div className="text-muted-foreground text-sm">Loading...</div>
        </div>
      )
    }

    // If we got here but not authenticated, we're redirecting
    if (!isAuthenticated) {
      return null
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

  // Wrap entire app in SidebarProvider so useSidebar is always available
  return (
    <SidebarProvider className="flex h-dvh flex-col overflow-hidden">
      {renderContent()}
    </SidebarProvider>
  )
}

export const Route = createRootRoute({ component: RootLayout })
