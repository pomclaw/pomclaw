import { gatewayFetch } from "@/api/gateway-http"

export interface SessionSummary {
  id: string
  title: string
  preview: string
  message_count: number
  created: string
  updated: string
}

export interface SessionDetail {
  id: string
  messages: {
    role: "user" | "assistant"
    content: string
    media?: string[]
  }[]
  summary: string
  created: string
  updated: string
}

export async function getSessions(
  agentId: string,
  offset: number = 0,
  limit: number = 20,
): Promise<SessionSummary[]> {
  const params = new URLSearchParams({
    agent_id: agentId,
    offset: offset.toString(),
    limit: limit.toString(),
  })

  const res = await gatewayFetch(`/api/sessions?${params.toString()}`)
  if (!res.ok) {
    throw new Error(`Failed to fetch sessions: ${res.status}`)
  }
  return res.json()
}

export async function getSessionHistory(agentId: string, id: string): Promise<SessionDetail> {
  const params = new URLSearchParams({
    agent_id: agentId,
  })
  const res = await gatewayFetch(`/api/sessions/${encodeURIComponent(id)}?${params.toString()}`)
  if (!res.ok) {
    throw new Error(`Failed to fetch session ${id}: ${res.status}`)
  }
  return res.json()
}

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

export async function deleteSession(agentId: string, id: string): Promise<void> {
  const params = new URLSearchParams({
    agent_id: agentId,
  })
  const res = await gatewayFetch(`/api/sessions/${encodeURIComponent(id)}?${params.toString()}`, {
    method: "DELETE",
  })
  if (!res.ok) {
    throw new Error(`Failed to delete session ${id}: ${res.status}`)
  }
}
