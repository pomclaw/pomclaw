import webapi from "./gocliRequest"
import * as components from "./pomclawComponents"
export * from "./pomclawComponents"

/**
 * @description "List all agents"
 */
export function listAgents() {
	return webapi.get<components.ListAgentsResp>(`/pomclaw-api/v1/agents`)
}

/**
 * @description "Create a new agent"
 * @param req
 */
export function createAgent(req: components.CreateAgentReq) {
	return webapi.post<components.CreateAgentResp>(`/pomclaw-api/v1/agents`, req)
}

/**
 * @description "List memory chunks for agent"
 * @param params
 */
export function listMemoryChunks(params: components.ListMemoryChunksReqParams, agentID: string) {
	return webapi.get<components.ListMemoryChunksResp>(`/pomclaw-api/v1/agents/${agentID}/memory/chunks`, params)
}

/**
 * @description "List memory documents for specific agent"
 * @param params
 */
export function getAgentMemoryDocuments(params: components.ListAgentMemoryDocumentsReqParams, agentID: string) {
	return webapi.get<components.ListMemoryDocumentsResp>(`/pomclaw-api/v1/agents/${agentID}/memory/documents`, params)
}

/**
 * @description "Get specific memory document by path"
 * @param params
 */
export function getMemoryDocument(params: components.GetMemoryDocumentReqParams, agentID: string, documentID: number) {
	return webapi.get<components.GetMemoryDocumentResp>(`/pomclaw-api/v1/agents/${agentID}/memory/documents/${documentID}`, params)
}

/**
 * @description "Create or update memory document"
 * @param params
 * @param req
 */
export function putMemoryDocument(params: components.PutMemoryDocumentReqParams, req: components.PutMemoryDocumentReq, agentID: string, documentID: number) {
	return webapi.put<components.PutMemoryDocumentResp>(`/pomclaw-api/v1/agents/${agentID}/memory/documents/${documentID}`, params, req)
}

/**
 * @description "Delete memory document"
 * @param params
 */
export function deleteMemoryDocument(params: components.DeleteMemoryDocumentReqParams, agentID: string, documentID: number) {
	return webapi.delete<components.DeleteMemoryDocumentResp>(`/pomclaw-api/v1/agents/${agentID}/memory/documents/${documentID}`, params)
}

/**
 * @description "Index single document"
 * @param params
 * @param req
 */
export function indexDocument(params: components.IndexDocumentReqParams, req: components.IndexDocumentReq, agentID: string) {
	return webapi.post<components.IndexDocumentResp>(`/pomclaw-api/v1/agents/${agentID}/memory/index`, params, req)
}

/**
 * @description "Index all documents"
 * @param params
 */
export function indexAll(params: components.IndexAllReqParams, agentID: string) {
	return webapi.post<components.IndexAllResp>(`/pomclaw-api/v1/agents/${agentID}/memory/index-all`, params)
}

/**
 * @description "Search memory documents"
 * @param params
 * @param req
 */
export function searchMemory(params: components.SearchMemoryReqParams, req: components.SearchMemoryReq, agentID: string) {
	return webapi.post<components.SearchMemoryResp>(`/pomclaw-api/v1/agents/${agentID}/memory/search`, params, req)
}

/**
 * @description "Get agent details"
 * @param params
 */
export function getAgent(params: components.GetAgentReqParams, agent_id: string) {
	return webapi.get<components.GetAgentResp>(`/pomclaw-api/v1/agents/${agent_id}`, params)
}

/**
 * @description "Update agent"
 * @param params
 * @param req
 */
export function updateAgent(params: components.UpdateAgentReqParams, req: components.UpdateAgentReq, agent_id: string) {
	return webapi.put<components.UpdateAgentResp>(`/pomclaw-api/v1/agents/${agent_id}`, params, req)
}

/**
 * @description "Delete agent"
 * @param params
 */
export function deleteAgent(params: components.DeleteAgentReqParams, agent_id: string) {
	return webapi.delete<components.DeleteAgentResp>(`/pomclaw-api/v1/agents/${agent_id}`, params)
}

/**
 * @description "Get agent authorizations (granted skills, MCP servers, tool policy)"
 * @param params
 */
export function getAgentAuthorizations(params: components.GetAgentAuthorizationsReqParams, agent_id: string) {
	return webapi.get<components.GetAgentAuthorizationsResp>(`/pomclaw-api/v1/agents/${agent_id}/authorizations`, params)
}

/**
 * @description "List agent bootstrap files (SOUL.md, AGENTS.md, etc.)"
 * @param params
 */
export function listAgentFiles(params: components.ListAgentFilesReqParams, agent_id: string) {
	return webapi.get<components.ListAgentFilesResp>(`/pomclaw-api/v1/agents/${agent_id}/files`, params)
}

/**
 * @description "Get specific agent bootstrap file content"
 * @param params
 */
export function getAgentFile(params: components.GetAgentFileReqParams, agent_id: string, name: string) {
	return webapi.get<components.GetAgentFileResp>(`/pomclaw-api/v1/agents/${agent_id}/files/${name}`, params)
}

/**
 * @description "Set agent bootstrap file content"
 * @param params
 * @param req
 */
export function setAgentFile(params: components.SetAgentFileReqParams, req: components.SetAgentFileReq, agent_id: string, name: string) {
	return webapi.put<components.SetAgentFileResp>(`/pomclaw-api/v1/agents/${agent_id}/files/${name}`, params, req)
}

/**
 * @description "List skills for agent with grant status"
 * @param params
 */
export function listAgentSkills(params: components.ListAgentSkillsReqParams, agent_id: string) {
	return webapi.get<components.SkillsWithGrantResp>(`/pomclaw-api/v1/agents/${agent_id}/skills`, params)
}

/**
 * @description "Get system prompt preview for agent"
 * @param params
 */
export function getSystemPromptPreview(params: components.GetSystemPromptPreviewReqParams, agent_id: string) {
	return webapi.get<components.GetSystemPromptPreviewResp>(`/pomclaw-api/v1/agents/${agent_id}/system-prompt-preview`, params)
}

/**
 * @description "Get cost summary by date, agent, model, or provider"
 * @param params
 */
export function getCostSummary(params: components.GetCostSummaryReqParams) {
	return webapi.get<components.GetCostSummaryResp>(`/pomclaw-api/v1/costs/summary`, params)
}

/**
 * @description "List all MCP servers granted to an agent"
 * @param params
 */
export function listAgentMCPServers(params: components.ListAgentMCPServersReqParams, agent_id: string) {
	return webapi.get<components.ListAgentMCPServersResp>(`/pomclaw-api/v1/mcp/grants/agent/${agent_id}`, params)
}

/**
 * @description "List all MCP servers"
 */
export function listMCPServers() {
	return webapi.get<components.ListMCPServersResp>(`/pomclaw-api/v1/mcp/servers`)
}

/**
 * @description "Create a new MCP server"
 * @param req
 */
export function createMCPServer(req: components.CreateMCPServerReq) {
	return webapi.post<components.CreateMCPServerResp>(`/pomclaw-api/v1/mcp/servers`, req)
}

/**
 * @description "Get MCP server details"
 * @param params
 */
export function getMCPServer(params: components.GetMCPServerReqParams, id: string) {
	return webapi.get<components.GetMCPServerResp>(`/pomclaw-api/v1/mcp/servers/${id}`, params)
}

/**
 * @description "Update MCP server"
 * @param params
 * @param req
 */
export function updateMCPServer(params: components.UpdateMCPServerReqParams, req: components.UpdateMCPServerReq, id: string) {
	return webapi.put<components.UpdateMCPServerResp>(`/pomclaw-api/v1/mcp/servers/${id}`, params, req)
}

/**
 * @description "Delete MCP server"
 * @param params
 */
export function deleteMCPServer(params: components.DeleteMCPServerReqParams, id: string) {
	return webapi.delete<components.DeleteMCPServerResp>(`/pomclaw-api/v1/mcp/servers/${id}`, params)
}

/**
 * @description "List all agent grants for an MCP server"
 * @param params
 */
export function listMCPServerGrants(params: components.ListMCPServerGrantsReqParams, id: string) {
	return webapi.get<components.ListMCPServerGrantsResp>(`/pomclaw-api/v1/mcp/servers/${id}/grants`, params)
}

/**
 * @description "Grant MCP server access to an agent"
 * @param params
 * @param req
 */
export function grantMCPServerAgent(params: components.GrantMCPServerAgentReqParams, req: components.GrantMCPServerAgentReq, id: string) {
	return webapi.post<components.GrantMCPServerAgentResp>(`/pomclaw-api/v1/mcp/servers/${id}/grants/agent`, params, req)
}

/**
 * @description "Revoke MCP server access from an agent"
 * @param params
 */
export function revokeMCPServerAgentGrant(params: components.RevokeMCPServerAgentGrantReqParams, id: string, agent_id: string) {
	return webapi.delete<components.RevokeMCPServerAgentGrantResp>(`/pomclaw-api/v1/mcp/servers/${id}/grants/agent/${agent_id}`, params)
}

/**
 * @description "Reconnect MCP server (evict connection pool)"
 * @param params
 */
export function reconnectMCPServer(params: components.ReconnectMCPServerReqParams, id: string) {
	return webapi.post<components.ReconnectMCPServerResp>(`/pomclaw-api/v1/mcp/servers/${id}/reconnect`, params)
}

/**
 * @description "List tools for an MCP server"
 * @param params
 */
export function listMCPServerTools(params: components.ListMCPServerToolsReqParams, id: string) {
	return webapi.get<components.ListMCPServerToolsResp>(`/pomclaw-api/v1/mcp/servers/${id}/tools`, params)
}

/**
 * @description "Test MCP server connection without saving"
 * @param req
 */
export function testMCPServerConnection(req: components.TestMCPServerConnectionReq) {
	return webapi.post<components.TestMCPServerConnectionResp>(`/pomclaw-api/v1/mcp/servers/test`, req)
}

/**
 * @description "List all memory documents (global)"
 */
export function listMemoryDocuments() {
	return webapi.get<components.ListMemoryDocumentsResp>(`/pomclaw-api/v1/memory/documents`)
}

/**
 * @description "List all providers"
 */
export function listProviders() {
	return webapi.get<components.ListProvidersResp>(`/pomclaw-api/v1/providers`)
}

/**
 * @description "Create a new provider"
 * @param req
 */
export function createProvider(req: components.CreateProviderReq) {
	return webapi.post<components.CreateProviderResp>(`/pomclaw-api/v1/providers`, req)
}

/**
 * @description "Get provider details"
 * @param params
 */
export function getProvider(params: components.GetProviderReqParams, id: number) {
	return webapi.get<components.GetProviderResp>(`/pomclaw-api/v1/providers/${id}`, params)
}

/**
 * @description "Update provider"
 * @param params
 * @param req
 */
export function updateProvider(params: components.UpdateProviderReqParams, req: components.UpdateProviderReq, id: number) {
	return webapi.put<components.UpdateProviderResp>(`/pomclaw-api/v1/providers/${id}`, params, req)
}

/**
 * @description "Delete provider"
 * @param params
 */
export function deleteProvider(params: components.DeleteProviderReqParams, id: number) {
	return webapi.delete<components.DeleteProviderResp>(`/pomclaw-api/v1/providers/${id}`, params)
}

/**
 * @description "List provider models"
 * @param params
 */
export function listProviderModels(params: components.ListProviderModelsReqParams, id: number) {
	return webapi.get<components.ListProviderModelsResp>(`/pomclaw-api/v1/providers/${id}/models`, params)
}

/**
 * @description "Verify provider"
 * @param params
 * @param req
 */
export function verifyProvider(params: components.VerifyProviderReqParams, req: components.VerifyProviderReq, id: number) {
	return webapi.post<components.VerifyProviderResp>(`/pomclaw-api/v1/providers/${id}/verify`, params, req)
}

/**
 * @description "List all sessions"
 * @param params
 */
export function listSessions(params: components.ListSessionsReqParams) {
	return webapi.get<components.ListSessionsResp>(`/pomclaw-api/v1/sessions`, params)
}

/**
 * @description "Create a new session"
 * @param req
 */
export function createSession(req: components.CreateSessionReq) {
	return webapi.post<components.CreateSessionResp>(`/pomclaw-api/v1/sessions`, req)
}

/**
 * @description "Get session details"
 * @param params
 */
export function getSession(params: components.GetSessionReqParams, id: number) {
	return webapi.get<components.GetSessionResp>(`/pomclaw-api/v1/sessions/${id}`, params)
}

/**
 * @description "Delete session"
 * @param params
 */
export function deleteSession(params: components.DeleteSessionReqParams, id: number) {
	return webapi.delete<components.DeleteSessionResp>(`/pomclaw-api/v1/sessions/${id}`, params)
}

/**
 * @description "List all skills"
 */
export function listSkills() {
	return webapi.get<components.SkillsResp>(`/pomclaw-api/v1/skills`)
}

/**
 * @description "Get skill details"
 * @param params
 */
export function getSkill(params: components.GetSkillReqParams, id: number) {
	return webapi.get<components.GetSkillResp>(`/pomclaw-api/v1/skills/${id}`, params)
}

/**
 * @description "Update skill"
 * @param params
 * @param req
 */
export function updateSkill(params: components.UpdateSkillReqParams, req: components.UpdateSkillReq, id: number) {
	return webapi.put<components.UpdateSkillResp>(`/pomclaw-api/v1/skills/${id}`, params, req)
}

/**
 * @description "List all agent grants for a skill"
 * @param params
 */
export function listSkillGrants(params: components.ListSkillGrantsReqParams, id: number) {
	return webapi.get<components.ListSkillGrantsResp>(`/pomclaw-api/v1/skills/${id}/grants`, params)
}

/**
 * @description "Grant skill to agent"
 * @param params
 * @param req
 */
export function grantSkillAgent(params: components.GrantSkillAgentReqParams, req: components.GrantSkillAgentReq, id: number) {
	return webapi.post<components.GrantSkillAgentResp>(`/pomclaw-api/v1/skills/${id}/grants/agent`, params, req)
}

/**
 * @description "Revoke skill grant from agent"
 * @param params
 */
export function revokeSkillAgentGrant(params: components.RevokeSkillAgentGrantReqParams, id: number, agent_id: string) {
	return webapi.delete<components.RevokeSkillAgentGrantResp>(`/pomclaw-api/v1/skills/${id}/grants/agent/${agent_id}`, params)
}

/**
 * @description "Upload skill zip package"
 */
export function uploadSkill() {
	return webapi.post<components.UploadSkillResp>(`/pomclaw-api/v1/skills/upload`)
}

/**
 * @description "Get system health status"
 */
export function getSystemHealth() {
	return webapi.get<components.GetSystemHealthResp>(`/pomclaw-api/v1/system/health`)
}

/**
 * @description "List all built-in tools"
 */
export function listBuiltinTools() {
	return webapi.get<components.ListBuiltinToolsResp>(`/pomclaw-api/v1/tools/builtin`)
}

/**
 * @description "Get built-in tool details"
 * @param params
 */
export function getBuiltinTool(params: components.GetBuiltinToolReqParams, name: string) {
	return webapi.get<components.GetBuiltinToolResp>(`/pomclaw-api/v1/tools/builtin/${name}`, params)
}

/**
 * @description "Update built-in tool (enabled, settings)"
 * @param params
 * @param req
 */
export function updateBuiltinTool(params: components.UpdateBuiltinToolReqParams, req: components.UpdateBuiltinToolReq, name: string) {
	return webapi.put<components.UpdateBuiltinToolResp>(`/pomclaw-api/v1/tools/builtin/${name}`, params, req)
}

/**
 * @description "List traces with optional filtering"
 * @param params
 */
export function listTraces(params: components.ListTracesReqParams) {
	return webapi.get<components.ListTracesResp>(`/pomclaw-api/v1/traces`, params)
}

/**
 * @description "Get trace details with spans"
 * @param params
 */
export function getTrace(params: components.GetTraceReqParams, traceID: string) {
	return webapi.get<components.GetTraceResp>(`/pomclaw-api/v1/traces/${traceID}`, params)
}

/**
 * @description "Export trace as gzip-compressed JSON"
 * @param params
 */
export function exportTrace(params: components.ExportTraceReqParams, traceID: string) {
	return webapi.get<components.ExportTraceResp>(`/pomclaw-api/v1/traces/${traceID}/export`, params)
}

/**
 * @description "Get usage summary"
 * @param params
 */
export function getUsageSummary(params: components.GetUsageSummaryReqParams) {
	return webapi.get<components.GetUsageSummaryResp>(`/pomclaw-api/v1/usage/summary`, params)
}

/**
 * @description "Get usage time series"
 * @param params
 */
export function getUsageTimeSeries(params: components.GetUsageTimeSeriesReqParams) {
	return webapi.get<components.GetUsageTimeSeriesResp>(`/pomclaw-api/v1/usage/timeseries`, params)
}
