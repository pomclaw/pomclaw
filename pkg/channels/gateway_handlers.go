package channels

import (
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/pkg/logger"
)

// handleChatSend 处理发送消息请求
func (g *GatewayChannel) handleChatSend(client *ClientConn, req *RequestFrame) {
	// 解析参数
	message, ok := req.Params["message"].(string)
	if !ok || message == "" {
		g.sendError(client, req.ID, "invalid_params", "message is required")
		return
	}

	// 支持 sessionKey (OpenClaw标准) 和 sessionId (向后兼容)
	sessionID, _ := req.Params["sessionKey"].(string)
	if sessionID == "" {
		sessionID, _ = req.Params["sessionId"].(string)
	}
	if sessionID == "" {
		// 创建新session
		sessionID = fmt.Sprintf("session_%s_%d", client.UserID, time.Now().UnixNano())
		logger.InfoCF("gateway", "Created new session", map[string]interface{}{
			"session_id": sessionID,
			"user_id":    client.UserID,
		})
	}

	// 获取或创建session信息
	agentID := g.getOrCreateSession(sessionID, client.UserID)

	logger.InfoCF("gateway", "Processing chat message", map[string]interface{}{
		"client_id":  client.ID,
		"session_id": sessionID,
		"agent_id":   agentID,
		"message":    message,
	})

	// 使用 BaseChannel 的 HandleMessage 方法发布到 MessageBus
	// 注意：ChatID 是 client.ID，这样 agent 响应时会根据这个 ID 路由回来
	metadata := map[string]string{
		"agent_id":   agentID,
		"req_id":     req.ID,
		"session_id": sessionID,
	}

	g.HandleMessage(client.UserID, client.ID, message, nil, metadata)

	// 立即返回响应确认
	g.sendResponse(client, req.ID, map[string]interface{}{
		"sessionId": sessionID,
		"status":    "processing",
	})
}

// handleSessionsList 处理查询会话列表
func (g *GatewayChannel) handleSessionsList(client *ClientConn, req *RequestFrame) {
	var sessions []map[string]interface{}

	// 遍历当前用户的sessions
	g.sessions.Range(func(key, value interface{}) bool {
		session := value.(*SessionInfo)
		if session.UserID == client.UserID {
			sessions = append(sessions, map[string]interface{}{
				"sessionId": session.SessionID,
				"agentId":   session.AgentID,
				"createdAt": session.CreatedAt.Unix(),
				"updatedAt": session.UpdatedAt.Unix(),
			})
		}
		return true
	})

	g.sendResponse(client, req.ID, map[string]interface{}{
		"sessions": sessions,
	})
}

// handleSessionsGet 处理获取单个会话
func (g *GatewayChannel) handleSessionsGet(client *ClientConn, req *RequestFrame) {
	sessionID, ok := req.Params["sessionId"].(string)
	if !ok || sessionID == "" {
		g.sendError(client, req.ID, "invalid_params", "sessionId is required")
		return
	}

	// 查询session
	val, ok := g.sessions.Load(sessionID)
	if !ok {
		g.sendError(client, req.ID, "session_not_found", "Session not found")
		return
	}

	session := val.(*SessionInfo)
	if session.UserID != client.UserID {
		g.sendError(client, req.ID, "permission_denied", "Access denied")
		return
	}

	g.sendResponse(client, req.ID, map[string]interface{}{
		"session": map[string]interface{}{
			"sessionId": session.SessionID,
			"agentId":   session.AgentID,
			"createdAt": session.CreatedAt.Unix(),
			"updatedAt": session.UpdatedAt.Unix(),
			"messages":  session.Messages,
		},
	})
}

// handleSessionsCreate 处理创建新会话
func (g *GatewayChannel) handleSessionsCreate(client *ClientConn, req *RequestFrame) {
	agentID, _ := req.Params["agentId"].(string)
	if agentID == "" {
		agentID = "default"
	}

	sessionID := fmt.Sprintf("session_%s_%d", client.UserID, time.Now().UnixNano())

	session := &SessionInfo{
		SessionID: sessionID,
		AgentID:   agentID,
		UserID:    client.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []Message{},
	}

	g.sessions.Store(sessionID, session)

	logger.InfoCF("gateway", "Session created", map[string]interface{}{
		"session_id": sessionID,
		"agent_id":   agentID,
		"user_id":    client.UserID,
	})

	// 发送响应
	g.sendResponse(client, req.ID, map[string]interface{}{
		"sessionId": sessionID,
		"agentId":   agentID,
	})

	// 发送事件
	g.sendEvent(client, "session.created", map[string]interface{}{
		"sessionId": sessionID,
		"agentId":   agentID,
	})
}

// handleSessionsDelete 处理删除会话
func (g *GatewayChannel) handleSessionsDelete(client *ClientConn, req *RequestFrame) {
	sessionID, ok := req.Params["sessionId"].(string)
	if !ok || sessionID == "" {
		g.sendError(client, req.ID, "invalid_params", "sessionId is required")
		return
	}

	// 检查权限
	val, ok := g.sessions.Load(sessionID)
	if !ok {
		g.sendError(client, req.ID, "session_not_found", "Session not found")
		return
	}

	session := val.(*SessionInfo)
	if session.UserID != client.UserID {
		g.sendError(client, req.ID, "permission_denied", "Access denied")
		return
	}

	// 删除session
	g.sessions.Delete(sessionID)

	logger.InfoCF("gateway", "Session deleted", map[string]interface{}{
		"session_id": sessionID,
		"user_id":    client.UserID,
	})

	g.sendResponse(client, req.ID, map[string]interface{}{
		"deleted": true,
	})
}

// getOrCreateSession 获取或创建session
func (g *GatewayChannel) getOrCreateSession(sessionID, userID string) string {
	// 查询session
	val, ok := g.sessions.Load(sessionID)
	if ok {
		session := val.(*SessionInfo)
		session.UpdatedAt = time.Now()
		return session.AgentID
	}

	// 创建新session
	agentID := "default"
	session := &SessionInfo{
		SessionID: sessionID,
		AgentID:   agentID,
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  []Message{},
	}

	g.sessions.Store(sessionID, session)

	logger.InfoCF("gateway", "Session initialized", map[string]interface{}{
		"session_id": sessionID,
		"agent_id":   agentID,
		"user_id":    userID,
	})

	return agentID
}
