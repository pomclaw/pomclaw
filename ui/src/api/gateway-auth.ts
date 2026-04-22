import { gatewayFetch } from "@/api/gateway-http"
import { getGatewayUrl } from "@/config/api"
import type { AuthUser } from "@/store/auth"

export interface AuthResponse {
  access_token: string
  user?: AuthUser
  user_id?: string
  username?: string
  email?: string
}

/**
 * Register a new user
 */
export async function registerUser(
  username: string,
  email: string,
  password: string,
): Promise<AuthResponse> {
  const res = await fetch(getGatewayUrl("/api/v1/auth/register"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, email, password }),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).error || `Registration failed: ${res.status}`,
    )
  }

  const data = (await res.json()) as AuthResponse
  // Convert backend response to expected format
  if (!data.user && data.user_id) {
    data.user = {
      id: data.user_id,
      username: data.username || "",
      email: email, // Use the email from the request since backend doesn't return it
    }
  }
  return data
}

/**
 * Login with username and password
 */
export async function loginUser(
  username: string,
  password: string,
): Promise<AuthResponse> {
  const res = await fetch(getGatewayUrl("/api/v1/auth/login"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password }),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).error || `Login failed: ${res.status}`,
    )
  }

  const data = (await res.json()) as AuthResponse
  // Convert backend response to expected format
  if (!data.user && data.user_id) {
    data.user = {
      id: data.user_id,
      username: data.username || "",
      email: data.email || "", // Backend doesn't return email on login
    }
  }
  return data
}

/**
 * Get current user info (requires valid token)
 */
export async function getCurrentUser(): Promise<AuthUser> {
  const res = await gatewayFetch("/api/v1/auth/me")

  if (!res.ok) {
    throw new Error(`Failed to fetch current user: ${res.status}`)
  }

  const data = (await res.json()) as { user: AuthUser }
  return data.user
}

/**
 * Logout (clears refresh token cookie)
 */
export async function logoutUser(): Promise<void> {
  try {
    await gatewayFetch("/api/v1/auth/logout", {
      method: "POST",
    })
  } catch {
    // Logout can fail if token is already invalid, but that's OK
    // We'll clear state anyway
  }
}

/**
 * Refresh access token
 */
export async function refreshAccessToken(): Promise<string> {
  const res = await fetch(getGatewayUrl("/api/v1/auth/refresh"), {
    method: "POST",
  })

  if (!res.ok) {
    throw new Error(`Token refresh failed: ${res.status}`)
  }

  const data = (await res.json()) as { access_token: string }
  return data.access_token
}
