package gateway

import (
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/logger"
)

// handleChatSend 处理发送消息请求
func (s *Server) handleChatSend(client *ClientConn, req *RequestFrame) {
	// 解析参数
	message, ok := req.Params["message"].(string)
	if !ok || message == "" {
		s.sendError(client, req.ID, "invalid_params", "message is required")
		return
	}

	sessionID, _ := req.Params["sessionId"].(string)
	if sessionID == "" {
		// 创建新session
		sessionID = fmt.Sprintf("session_%s_%d", client.UserID, time.Now().UnixNano())
		logger.InfoCF("gateway", "Created new session", map[string]interface{}{
			"session_id": sessionID,
			"user_id":    client.UserID,
		})
	}

	// 获取或创建session信息
	agentID := s.getOrCreateSession(sessionID, client.UserID)

	logger.InfoCF("gateway", "Processing chat message", map[string]interface{}{
		"client_id":  client.ID,
		"session_id": sessionID,
		"agent_id":   agentID,
		"message":    message,
	})

	// 构造InboundMessage发送到MessageBus
	inboundMsg := bus.InboundMessage{
		Channel:    "gateway",
		SenderID:   client.UserID,
		ChatID:     client.ID,
		Content:    message,
		SessionKey: sessionID,
		Metadata: map[string]string{
			"agent_id": agentID,
			"req_id":   req.ID,
		},
	}

	// 发布到MessageBus，由AgentLoop异步处理
	s.bus.PublishInbound(inboundMsg)

	// 立即返回响应确认
	s.sendResponse(client, req.ID, map[string]interface{}{
		"sessionId": sessionID,
		"status":    "processing",
	})
}

// handleSessionsList 处理查询会话列表
func (s *Server) handleSessionsList(client *ClientConn, req *RequestFrame) {
	var sessions []map[string]interface{}

	// 遍历当前用户的sessions
	s.sessions.Range(func(key, value interface{}) bool {
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

	s.sendResponse(client, req.ID, map[string]interface{}{
		"sessions": sessions,
	})
}

// handleSessionsGet 处理获取单个会话
func (s *Server) handleSessionsGet(client *ClientConn, req *RequestFrame) {
	sessionID, ok := req.Params["sessionId"].(string)
	if !ok || sessionID == "" {
		s.sendError(client, req.ID, "invalid_params", "sessionId is required")
		return
	}

	// 查询session
	val, ok := s.sessions.Load(sessionID)
	if !ok {
		s.sendError(client, req.ID, "session_not_found", "Session not found")
		return
	}

	session := val.(*SessionInfo)
	if session.UserID != client.UserID {
		s.sendError(client, req.ID, "permission_denied", "Access denied")
		return
	}

	s.sendResponse(client, req.ID, map[string]interface{}{
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
func (s *Server) handleSessionsCreate(client *ClientConn, req *RequestFrame) {
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

	s.sessions.Store(sessionID, session)

	logger.InfoCF("gateway", "Session created", map[string]interface{}{
		"session_id": sessionID,
		"agent_id":   agentID,
		"user_id":    client.UserID,
	})

	// 发送响应
	s.sendResponse(client, req.ID, map[string]interface{}{
		"sessionId": sessionID,
		"agentId":   agentID,
	})

	// 发送事件
	s.sendEvent(client, "session.created", map[string]interface{}{
		"sessionId": sessionID,
		"agentId":   agentID,
	})
}

// handleSessionsDelete 处理删除会话
func (s *Server) handleSessionsDelete(client *ClientConn, req *RequestFrame) {
	sessionID, ok := req.Params["sessionId"].(string)
	if !ok || sessionID == "" {
		s.sendError(client, req.ID, "invalid_params", "sessionId is required")
		return
	}

	// 检查权限
	val, ok := s.sessions.Load(sessionID)
	if !ok {
		s.sendError(client, req.ID, "session_not_found", "Session not found")
		return
	}

	session := val.(*SessionInfo)
	if session.UserID != client.UserID {
		s.sendError(client, req.ID, "permission_denied", "Access denied")
		return
	}

	// 删除session
	s.sessions.Delete(sessionID)

	logger.InfoCF("gateway", "Session deleted", map[string]interface{}{
		"session_id": sessionID,
		"user_id":    client.UserID,
	})

	s.sendResponse(client, req.ID, map[string]interface{}{
		"deleted": true,
	})
}

// getOrCreateSession 获取或创建session
func (s *Server) getOrCreateSession(sessionID, userID string) string {
	// 查询session
	val, ok := s.sessions.Load(sessionID)
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

	s.sessions.Store(sessionID, session)

	logger.InfoCF("gateway", "Session initialized", map[string]interface{}{
		"session_id": sessionID,
		"agent_id":   agentID,
		"user_id":    userID,
	})

	return agentID
}
