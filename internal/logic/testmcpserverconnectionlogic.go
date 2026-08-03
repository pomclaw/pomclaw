package logic

import (
	"context"
	"encoding/json"

	"github.com/pomclaw/pomclaw/internal/mcp"
	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type TestMCPServerConnectionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Test MCP server connection without saving
func NewTestMCPServerConnectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestMCPServerConnectionLogic {
	return &TestMCPServerConnectionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestMCPServerConnectionLogic) TestMCPServerConnection(req *types.TestMCPServerConnectionReq) (resp *types.TestMCPServerConnectionResp, err error) {
	_, err = GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("TestMCPServerConnection failed: %v", err)
		return nil, err
	}

	if req.Transport == "" {
		return nil, &ValidationError{Message: "transport is required"}
	}

	var args []string
	if req.Args != "" {
		_ = json.Unmarshal([]byte(req.Args), &args)
	}

	var env map[string]string
	if req.Env != "" {
		_ = json.Unmarshal([]byte(req.Env), &env)
	}

	var headers map[string]string
	if req.Headers != "" {
		_ = json.Unmarshal([]byte(req.Headers), &headers)
	}

	tools, err := mcp.DiscoverTools(l.ctx, req.Transport, req.Command, args, env, req.URL, headers)
	if err != nil {
		return &types.TestMCPServerConnectionResp{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &types.TestMCPServerConnectionResp{
		Success:   true,
		ToolCount: len(tools),
	}, nil
}