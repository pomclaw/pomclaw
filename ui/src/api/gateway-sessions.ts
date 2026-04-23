import { gatewayFetch } from "@/api/gateway-http"

export interface ChatMessage {
  role: "user" | "assistant"
  content: string
  media?: string[]
}

export interface SessionListItem {
  id: string
  title: string
  preview: string
  message_count: number
  created: string
  updated: string
}

export interface SessionDetail {
  id: string
  messages: ChatMessage[]
  summary: string
  created: string
  updated: string
}

/**
 * List all sessions with pagination
 */
export async function listSessions(
  offset: number = 0,
  limit: number = 20,
): Promise<SessionListItem[]> {
  const params = new URLSearchParams({
    offset: offset.toString(),
    limit: limit.toString(),
  })

  const res = await gatewayFetch(`/api/sessions?${params.toString()}`)

  if (!res.ok) {
    throw new Error(`Failed to list sessions: ${res.status}`)
  }

  return res.json()
}

/**
 * Get a single session with messages by ID
 */
export async function getSession(sessionId: string): Promise<SessionDetail> {
  const res = await gatewayFetch(`/api/sessions/${encodeURIComponent(sessionId)}`)

  if (!res.ok) {
    throw new Error(`Failed to get session: ${res.status}`)
  }

  return res.json()
}

/**
 * Create a new session
 */
export async function createSession(
  agentId: string,
  title?: string,
): Promise<SessionDetail> {
  const payload = {
    agent_id: agentId,
    title: title || "",
  }

  const res = await gatewayFetch("/api/sessions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).message || `Failed to create session: ${res.status}`,
    )
  }

  return res.json()
}

/**
 * Delete a session
 */
export async function deleteSession(sessionId: string): Promise<void> {
  const res = await gatewayFetch(`/api/sessions/${encodeURIComponent(sessionId)}`, {
    method: "DELETE",
  })

  if (!res.ok) {
    throw new Error(`Failed to delete session: ${res.status}`)
  }
}
