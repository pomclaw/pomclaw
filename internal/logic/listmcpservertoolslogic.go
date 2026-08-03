package logic

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/pomclaw/pomclaw/internal/mcp"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListMCPServerToolsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List tools for an MCP server
func NewListMCPServerToolsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMCPServerToolsLogic {
	return &ListMCPServerToolsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMCPServerToolsLogic) ListMCPServerTools(req *types.ListMCPServerToolsReq) (resp *types.ListMCPServerToolsResp, err error) {
	userID, err := GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("ListMCPServerTools failed: %v", err)
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
		l.Errorf("ListMCPServerTools failed: %v", err)
		return nil, err
	}

	if server.UserId != userID {
		return nil, &NotFoundError{Message: "MCP server not found"}
	}

	var args []string
	if server.Args != "" {
		_ = json.Unmarshal([]byte(server.Args), &args)
	}

	var env map[string]string
	if server.Env != "" {
		_ = json.Unmarshal([]byte(server.Env), &env)
	}

	var headers map[string]string
	if server.Headers != "" {
		_ = json.Unmarshal([]byte(server.Headers), &headers)
	}

	command := ""
	if server.Command.Valid {
		command = server.Command.String
	}
	url := ""
	if server.Url.Valid {
		url = server.Url.String
	}

	tools, err := mcp.DiscoverTools(l.ctx, server.Transport, command, args, env, url, headers)
	if err != nil {
		l.Errorf("ListMCPServerTools failed to discover tools: %v", err)
		return &types.ListMCPServerToolsResp{
			Tools: []types.MCPToolInfo{},
		}, nil
	}

	result := make([]types.MCPToolInfo, 0, len(tools))
	for _, t := range tools {
		result = append(result, types.MCPToolInfo{
			Name:        t.Name,
			Description: t.Description,
		})
	}

	return &types.ListMCPServerToolsResp{
		Tools: result,
	}, nil
}