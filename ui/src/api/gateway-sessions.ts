import { gatewayFetch } from "@/api/gateway-http"

export interface ChatMessage {
  id: string
  role: "user" | "assistant"
  content: string
  created_at?: string
}

export interface Session {
  id: string
  agent_id: string
  title?: string
  status?: string
  created_at?: string
  updated_at?: string
  message_count?: number
}

export interface SessionListResponse {
  sessions: Session[]
  total: number
  page: number
  page_size: number
}

// Backend response format (Go struct with uppercase fields)
interface BackendSession {
  ID: string
  AgentID: string
  Title?: string
  Status?: string
  MessageCount?: number
  CreatedAt?: string
  UpdatedAt?: string
}

function mapBackendSession(backendSession: BackendSession): Session {
  return {
    id: backendSession.ID,
    agent_id: backendSession.AgentID,
    title: backendSession.Title,
    status: backendSession.Status,
    message_count: backendSession.MessageCount,
    created_at: backendSession.CreatedAt,
    updated_at: backendSession.UpdatedAt,
  }
}

export interface SessionChatHistoryResponse {
  messages: ChatMessage[]
  total: number
  page: number
  page_size: number
}

/**
 * List all sessions for a specific agent
 */
export async function listSessions(
  agentId: string,
  page: number = 1,
  pageSize: number = 50,
): Promise<SessionListResponse> {
  const params = new URLSearchParams({
    agent_id: agentId,
    page: page.toString(),
    page_size: pageSize.toString(),
  })

  const res = await gatewayFetch(`/api/v1/sessions?${params.toString()}`)

  if (!res.ok) {
    throw new Error(`Failed to list sessions: ${res.status}`)
  }

  const data = await res.json()

  // Backend returns array directly, transform to expected format
  const sessions = Array.isArray(data) ? data.map(mapBackendSession) : (data.sessions || []).map(mapBackendSession)

  return {
    sessions,
    total: sessions.length,
    page,
    page_size: pageSize,
  }
}

/**
 * Get a single session by ID
 */
export async function getSession(sessionId: string): Promise<Session> {
  const res = await gatewayFetch(`/api/v1/sessions/${encodeURIComponent(sessionId)}`)

  if (!res.ok) {
    throw new Error(`Failed to get session: ${res.status}`)
  }

  const data = await res.json()
  return mapBackendSession(data)
}

/**
 * Create a new session
 */
export async function createSession(
  agentId: string,
  title?: string,
): Promise<Session> {
  const backendPayload = {
    agent_id: agentId,
    title: title,
  }

  const res = await gatewayFetch("/api/v1/sessions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(backendPayload),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).error || `Failed to create session: ${res.status}`,
    )
  }

  const data = await res.json()
  return mapBackendSession(data)
}

/**
 * Update an existing session
 */
export async function updateSession(
  sessionId: string,
  updates: Partial<Session>,
): Promise<Session> {
  // Convert frontend format to backend format
  const backendUpdates: Record<string, unknown> = {}
  if (updates.title !== undefined) backendUpdates.Title = updates.title
  if (updates.status !== undefined) backendUpdates.Status = updates.status

  const res = await gatewayFetch(`/api/v1/sessions/${encodeURIComponent(sessionId)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(backendUpdates),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).error || `Failed to update session: ${res.status}`,
    )
  }

  const data = await res.json()
  return mapBackendSession(data)
}

/**
 * Delete a session
 */
export async function deleteSession(sessionId: string): Promise<void> {
  const res = await gatewayFetch(`/api/v1/sessions/${encodeURIComponent(sessionId)}`, {
    method: "DELETE",
  })

  if (!res.ok) {
    throw new Error(`Failed to delete session: ${res.status}`)
  }
}

/**
 * Load chat history for a session
 */
export async function loadSessionChatHistory(
  sessionId: string,
  page: number = 1,
  pageSize: number = 50,
): Promise<SessionChatHistoryResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  })

  const res = await gatewayFetch(
    `/api/v1/sessions/${encodeURIComponent(sessionId)}/history?${params.toString()}`,
  )

  if (!res.ok) {
    throw new Error(`Failed to load chat history: ${res.status}`)
  }

  return res.json()
}
