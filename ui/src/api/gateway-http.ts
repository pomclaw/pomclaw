import { getDefaultStore } from "jotai"

import { getGatewayUrl } from "@/config/api"
import { accessTokenAtom, updateAuthStore } from "@/store/auth"

function isAuthPage(): boolean {
  if (typeof globalThis.location === "undefined") {
    return false
  }
  const pathname = globalThis.location.pathname || "/"
  return pathname === "/login" || pathname === "/register"
}

/**
 * Fetch wrapper for Gateway API that:
 * - Adds JWT bearer token in Authorization header
 * - Auto-refreshes on 401 (if token is stale)
 * - Redirects to /login on permanent auth failure
 * - Uses credentials: "include" for refresh token cookie
 */
export async function gatewayFetch(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<Response> {
  const store = getDefaultStore()
  const accessToken = store.get(accessTokenAtom)

  // Convert relative URLs to absolute if needed
  let url = input instanceof URL ? input.toString() : String(input)
  if (typeof url === "string" && url.startsWith("/")) {
    url = getGatewayUrl(url)
  }

  const headers = new Headers(init?.headers || {})
  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`)
  }

  let res = await fetch(url, {
    ...init,
    headers,
  })

  // If 401 and we have a token, try to refresh
  if (res.status === 401 && accessToken) {
    const refreshRes = await fetch(getGatewayUrl("/api/v1/auth/refresh"), {
      method: "POST",
    })

    if (refreshRes.ok) {
      const data = (await refreshRes.json()) as { access_token: string }
      updateAuthStore((prev) => ({
        ...prev,
        accessToken: data.access_token,
      }))

      // Retry the original request with new token
      headers.set("Authorization", `Bearer ${data.access_token}`)
      res = await fetch(url, {
        ...init,
        headers,
      })
    } else {
      // Refresh failed - permanently unauthenticated
      if (
        typeof globalThis.location !== "undefined" &&
        !isAuthPage()
      ) {
        globalThis.location.assign("/login")
      }
      updateAuthStore((prev) => ({
        ...prev,
        accessToken: null,
        user: null,
        isAuthenticated: false,
        error: "Session expired. Please log in again.",
      }))
    }
  }

  return res
}
