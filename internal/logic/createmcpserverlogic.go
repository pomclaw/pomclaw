package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateMCPServerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create a new MCP server
func NewCreateMCPServerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMCPServerLogic {
	return &CreateMCPServerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateMCPServerLogic) CreateMCPServer(req *types.CreateMCPServerReq) (resp *types.CreateMCPServerResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("CreateMCPServer failed: %v", err)
		return nil, err
	}

	if req.Name == "" {
		return nil, &ValidationError{Message: "name is required"}
	}
	if req.Transport == "" {
		return nil, &ValidationError{Message: "transport is required"}
	}
	if req.Transport != "stdio" && req.Transport != "sse" && req.Transport != "streamable-http" {
		return nil, &ValidationError{Message: "transport must be stdio, sse, or streamable-http"}
	}
	if req.Transport == "stdio" && req.Command == "" {
		return nil, &ValidationError{Message: "command is required for stdio transport"}
	}
	if req.Transport != "stdio" && req.URL == "" {
		return nil, &ValidationError{Message: "url is required for sse/streamable-http transport"}
	}

	// Check name uniqueness
	existing, err := l.svcCtx.McpServersModel.FindByUserIDName(l.ctx, userID, req.Name)
	if err != nil && err != model.ErrNotFound {
		l.Errorf("CreateMCPServer failed: %v", err)
		return nil, err
	}
	if existing != nil {
		return nil, &ValidationError{Message: fmt.Sprintf("MCP server with name '%s' already exists", req.Name)}
	}

	now := time.Now()
	settings := req.Settings
	if settings == "" {
		settings = "{}"
	}
	args := req.Args
	if args == "" {
		args = "[]"
	}
	headers := req.Headers
	if headers == "" {
		headers = "{}"
	}
	env := req.Env
	if env == "" {
		env = "{}"
	}

	p := &model.McpServers{
		UserId:      userID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Transport:   req.Transport,
		Command:     sql.NullString{String: req.Command, Valid: req.Command != ""},
		Args:        args,
		Url:         sql.NullString{String: req.URL, Valid: req.URL != ""},
		Headers:     headers,
		Env:         env,
		ApiKey:      sql.NullString{String: req.APIKey, Valid: req.APIKey != ""},
		ToolPrefix:  sql.NullString{String: req.ToolPrefix, Valid: req.ToolPrefix != ""},
		TimeoutSec:  int64(req.TimeoutSec),
		Settings:    settings,
		Enabled:     req.Enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = l.svcCtx.McpServersModel.Insert(l.ctx, p)
	if err != nil {
		l.Errorf("CreateMCPServer failed: %v", err)
		return nil, err
	}

	// Fetch the created record to get the generated ID
	created, err := l.svcCtx.McpServersModel.FindByUserIDName(l.ctx, userID, req.Name)
	if err != nil {
		l.Errorf("CreateMCPServer failed to read back: %v", err)
		return nil, err
	}

	server := convertModelMCPServerToType(created)
	if user, err := l.svcCtx.UsersModel.FindOneByUserId(l.ctx, created.UserId); err == nil {
		server.CreatedByName = user.Username
	}
	return &types.CreateMCPServerResp{
		Server: server,
	}, nil
}
