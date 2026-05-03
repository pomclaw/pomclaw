import type { ReactNode } from "react"
import { useEffect } from "react"

import { useHighlightTheme } from "./hooks/use-highlight-theme"
import { initWsClient } from "@/lib/ws-client"

interface AppProvidersProps {
  children: ReactNode
}

export function AppProviders({ children }: AppProvidersProps) {
  useHighlightTheme()

  // Initialize WebSocket client on mount
  useEffect(() => {
    const wsPort = import.meta.env.VITE_API_WEBSOCKET_PORT || '18792'
    const wsUrl = `ws://${window.location.hostname}:${wsPort}/ws`
    const userId = 'user-' + Math.random().toString(36).substring(7) // TODO: Real user ID from auth

    initWsClient({ url: wsUrl, userId })
  }, [])

  return <>{children}</>
}
