/**
 * API Configuration
 *
 * Get configuration from:
 * 1. Environment variables (VITE_API_* during build)
 * 2. Runtime window config (window.pomclawConfig)
 * 3. Auto-detect from window.location
 */

interface ApiConfig {
  /** Base URL for Gateway API calls (e.g., http://localhost:8080) */
  gatewayBaseUrl: string
  /** Base URL for WebSocket connections (e.g., ws://localhost:18792) */
  websocketBaseUrl: string
}

/**
 * Get API configuration
 *
 * @example
 * // Uses VITE_API_GATEWAY_PORT env var if set during build
 * const config = getApiConfig()
 *
 * // Or at runtime, set window.pomclawConfig before app loads:
 * window.pomclawConfig = {
 *   gatewayPort: 18792,  // API 和 WebSocket 使用同一个端口
 * }
 */
export function getApiConfig(): ApiConfig {
  // Runtime config takes highest priority
  const runtimeConfig = (globalThis as any).pomclawConfig as {
    gatewayPort?: number
  } | undefined

  if (runtimeConfig?.gatewayPort) {
    const scheme = window.location.protocol === "https:" ? "https" : "http"
    const wsScheme = window.location.protocol === "https:" ? "wss" : "ws"
    const port = runtimeConfig.gatewayPort

    return {
      gatewayBaseUrl: `${scheme}://${window.location.hostname}:${port}`,
      websocketBaseUrl: `${wsScheme}://${window.location.hostname}:${port}`,
    }
  }

  // Build-time environment variable
  const gatewayPort = import.meta.env.VITE_API_GATEWAY_PORT as string | undefined

  if (gatewayPort) {
    const scheme = window.location.protocol === "https:" ? "https" : "http"
    const wsScheme = window.location.protocol === "https:" ? "wss" : "ws"

    return {
      gatewayBaseUrl: `${scheme}://${window.location.hostname}:${gatewayPort}`,
      websocketBaseUrl: `${wsScheme}://${window.location.hostname}:${gatewayPort}`,
    }
  }

  // Default: assume both APIs are on same server
  const wsScheme = window.location.protocol === "https:" ? "wss" : "ws"

  return {
    gatewayBaseUrl: window.location.origin,
    websocketBaseUrl: `${wsScheme}://${window.location.host}`,
  }
}

/**
 * Build a full Gateway API URL
 * @example
 * getGatewayUrl("/api/v1/agents") // http://localhost:8080/api/v1/agents
 */
export function getGatewayUrl(path: string): string {
  const config = getApiConfig()
  return `${config.gatewayBaseUrl}${path}`
}

/**
 * Build a WebSocket URL
 * @example
 * getWebSocketUrl("/ws") // ws://localhost:18792/ws
 */
export function getWebSocketUrl(path: string): string {
  const config = getApiConfig()
  return `${config.websocketBaseUrl}${path}`
}
