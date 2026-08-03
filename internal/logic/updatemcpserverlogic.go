package logic

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMCPServerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Update MCP server
func NewUpdateMCPServerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMCPServerLogic {
	return &UpdateMCPServerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMCPServerLogic) UpdateMCPServer(req *types.UpdateMCPServerReq) (resp *types.UpdateMCPServerResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("UpdateMCPServer failed: %v", err)
		return nil, err
	}

	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, &ValidationError{Message: "invalid server id"}
	}

	server, err := l.svcCtx.McpServersModel.FindOne(l.ctx, id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, &NotFoundError{Message: "MCP server not found"}
		}
		l.Errorf("UpdateMCPServer failed: %v", err)
		return nil, err
	}

	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Description != "" {
		server.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.Transport != "" {
		server.Transport = req.Transport
	}
	if req.Command != "" {
		server.Command = sql.NullString{String: req.Command, Valid: true}
	}
	if req.Args != "" {
		server.Args = req.Args
	}
	if req.URL != "" {
		server.Url = sql.NullString{String: req.URL, Valid: true}
	}
	if req.Headers != "" {
		server.Headers = req.Headers
	}
	if req.Env != "" {
		server.Env = req.Env
	}
	if req.APIKey != "" && req.APIKey != "***" {
		server.ApiKey = sql.NullString{String: req.APIKey, Valid: true}
	}
	if req.ToolPrefix != "" {
		server.ToolPrefix = sql.NullString{String: req.ToolPrefix, Valid: true}
	}
	if req.TimeoutSec != 0 {
		server.TimeoutSec = int64(req.TimeoutSec)
	}
	if req.Settings != "" {
		server.Settings = req.Settings
	}
	server.Enabled = req.Enabled
	server.UpdatedAt = time.Now()

	err = l.svcCtx.McpServersModel.Update(l.ctx, server)
	if err != nil {
		l.Errorf("UpdateMCPServer failed: %v", err)
		return nil, err
	}

	result := convertModelMCPServerToType(server)
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, server.UserId); err == nil {
		result.CreatedByName = user.Username
	}
	return &types.UpdateMCPServerResp{
		Server: result,
	}, nil
}