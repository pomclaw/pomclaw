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
 * Configure via .env files:
 * - VITE_API_GATEWAY_PORT: Gateway REST API port (default: 18792)
 * - VITE_API_WEBSOCKET_PORT: WebSocket port (default: 18792)
 * - VITE_API_HOSTNAME: Override hostname (optional, default: localhost)
 */
export function getApiConfig(): ApiConfig {
  const scheme = window.location.protocol === "https:" ? "https" : "http"
  const wsScheme = window.location.protocol === "https:" ? "wss" : "ws"

  // Read from environment variables (from .env files)
  const gatewayPort = import.meta.env.VITE_API_GATEWAY_PORT || "18792"
  const websocketPort = import.meta.env.VITE_API_WEBSOCKET_PORT || "18792"
  const hostname = import.meta.env.VITE_API_HOSTNAME || window.location.hostname

  return {
    gatewayBaseUrl: `${scheme}://${hostname}:${gatewayPort}`,
    websocketBaseUrl: `${wsScheme}://${hostname}:${websocketPort}`,
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
