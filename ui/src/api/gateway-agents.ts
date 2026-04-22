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

// Backend response format (Go struct with uppercase fields)
interface BackendAgent {
  ID: string
  UserID: string
  Name: string
  Description?: string
  SystemPrompt?: string
  Model?: string
  Tools?: string[]
  Status?: string
  CreatedAt?: string
  UpdatedAt?: string
}

function mapBackendAgent(backendAgent: BackendAgent): Agent {
  return {
    id: backendAgent.ID,
    name: backendAgent.Name,
    description: backendAgent.Description,
    system_prompt: backendAgent.SystemPrompt,
    model: backendAgent.Model,
    tools: backendAgent.Tools || [],
    created_at: backendAgent.CreatedAt,
    updated_at: backendAgent.UpdatedAt,
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

  // Backend returns array directly, transform to expected format
  const agents = Array.isArray(data) ? data.map(mapBackendAgent) : (data.agents || []).map(mapBackendAgent)

  return {
    agents,
    total: agents.length,
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
  // Convert frontend format to backend format
  const backendAgent = {
    Name: agent.name,
    Description: agent.description,
    SystemPrompt: agent.system_prompt,
    Model: agent.model,
    Tools: agent.tools,
  }

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
  // Convert frontend format to backend format
  const backendUpdates: Record<string, unknown> = {}
  if (updates.name !== undefined) backendUpdates.Name = updates.name
  if (updates.description !== undefined) backendUpdates.Description = updates.description
  if (updates.system_prompt !== undefined) backendUpdates.SystemPrompt = updates.system_prompt
  if (updates.model !== undefined) backendUpdates.Model = updates.model
  if (updates.tools !== undefined) backendUpdates.Tools = updates.tools

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
