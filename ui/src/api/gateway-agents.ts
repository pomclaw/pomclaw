import { gatewayFetch } from "@/api/gateway-http"

export interface Agent {
  id: string
  name: string
  description?: string
  system_prompt?: string
  model?: string
  tools?: string[]
  created_at?: string
  updated_at?: string
}

export interface AgentListResponse {
  agents: Agent[]
  total: number
  page: number
  page_size: number
}

// Backend response format (lowercase snake_case from Go JSON tags)
interface BackendAgent {
  id: string
  user_id: string
  name: string
  description?: string
  system_prompt?: string
  model?: string
  tools?: string[]
  status?: string
  created_at?: string
  updated_at?: string
}

function mapBackendAgent(backendAgent: BackendAgent): Agent {
  return {
    id: backendAgent.id,
    name: backendAgent.name,
    description: backendAgent.description,
    system_prompt: backendAgent.system_prompt,
    model: backendAgent.model,
    tools: backendAgent.tools || [],
    created_at: backendAgent.created_at,
    updated_at: backendAgent.updated_at,
  }
}

/**
 * List all agents for the current user
 */
export async function listAgents(
  page: number = 1,
  pageSize: number = 50,
): Promise<AgentListResponse> {
  const params = new URLSearchParams({
    page: page.toString(),
    page_size: pageSize.toString(),
  })

  const res = await gatewayFetch(`/api/v1/agents?${params.toString()}`)

  if (!res.ok) {
    throw new Error(`Failed to list agents: ${res.status}`)
  }

  const data = await res.json()

  // Backend returns { total, agents } format
  const agents = (data.agents || []).map(mapBackendAgent)

  return {
    agents,
    total: data.total || agents.length,
    page,
    page_size: pageSize,
  }
}

/**
 * Get a single agent by ID
 */
export async function getAgent(agentId: string): Promise<Agent> {
  const res = await gatewayFetch(`/api/v1/agents/${encodeURIComponent(agentId)}`)

  if (!res.ok) {
    throw new Error(`Failed to get agent: ${res.status}`)
  }

  const data = await res.json()
  return mapBackendAgent(data)
}

/**
 * Create a new agent
 */
export async function createAgent(agent: Partial<Agent>): Promise<Agent> {
  // Send with lowercase field names to match backend expectations
  // Filter out undefined fields
  const backendAgent: Record<string, unknown> = {
    name: agent.name,
    model: agent.model,
  }

  if (agent.description !== undefined) backendAgent.description = agent.description
  if (agent.system_prompt !== undefined) backendAgent.system_prompt = agent.system_prompt
  if (agent.tools !== undefined) backendAgent.tools = agent.tools

  const res = await gatewayFetch("/api/v1/agents", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(backendAgent),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).error || `Failed to create agent: ${res.status}`,
    )
  }

  const data = await res.json()
  return mapBackendAgent(data)
}

/**
 * Update an existing agent
 */
export async function updateAgent(agentId: string, updates: Partial<Agent>): Promise<Agent> {
  // Send with lowercase field names to match backend expectations
  const backendUpdates: Record<string, unknown> = {}
  if (updates.name !== undefined) backendUpdates.name = updates.name
  if (updates.description !== undefined) backendUpdates.description = updates.description
  if (updates.system_prompt !== undefined) backendUpdates.system_prompt = updates.system_prompt
  if (updates.model !== undefined) backendUpdates.model = updates.model
  if (updates.tools !== undefined) backendUpdates.tools = updates.tools

  const res = await gatewayFetch(`/api/v1/agents/${encodeURIComponent(agentId)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(backendUpdates),
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({}))
    throw new Error(
      (error as Record<string, string>).error || `Failed to update agent: ${res.status}`,
    )
  }

  const data = await res.json()
  return mapBackendAgent(data)
}

/**
 * Delete an agent
 */
export async function deleteAgent(agentId: string): Promise<void> {
  const res = await gatewayFetch(`/api/v1/agents/${encodeURIComponent(agentId)}`, {
    method: "DELETE",
  })

  if (!res.ok) {
    throw new Error(`Failed to delete agent: ${res.status}`)
  }
}
