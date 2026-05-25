package logic

import (
	"context"

	"github.com/pomclaw/pomclaw/internal/svc"
	"github.com/pomclaw/pomclaw/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListBuiltinToolsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBuiltinToolsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBuiltinToolsLogic {
	return &ListBuiltinToolsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ListBuiltinTools lists all built-in tools
func (l *ListBuiltinToolsLogic) ListBuiltinTools() (*types.ListBuiltinToolsResp, error) {

	return &types.ListBuiltinToolsResp{
		Tools: []types.BuiltinToolDef{
			{
				Name:     "test",
				Display:  "testtest",
				Desc:     "desc",
				Enabled:  true,
				Settings: nil,
			},
		},
	}, nil
}

type GetBuiltinToolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBuiltinToolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBuiltinToolLogic {
	return &GetBuiltinToolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetBuiltinTool gets a built-in tool by name
func (l *GetBuiltinToolLogic) GetBuiltinTool(req *types.GetBuiltinToolReq) (*types.BuiltinToolDef, error) {
	l.Infof("GetBuiltinTool called with name: %s", req.Name)

	// TODO: Implement getting builtin tool from storage/service
	return &types.BuiltinToolDef{
		Name:    req.Name,
		Enabled: false,
	}, nil
}

type UpdateBuiltinToolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBuiltinToolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBuiltinToolLogic {
	return &UpdateBuiltinToolLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateBuiltinTool updates a built-in tool settings
func (l *UpdateBuiltinToolLogic) UpdateBuiltinTool(req *types.UpdateBuiltinToolReq) (*types.UpdateBuiltinToolResp, error) {
	l.Infof("UpdateBuiltinTool called with name: %s", req.Name)

	// TODO: Implement updating builtin tool in storage/service
	return &types.UpdateBuiltinToolResp{
		Status: "updated",
	}, nil
}

type GetTenantConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantConfigLogic {
	return &GetTenantConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTenantConfig gets tenant-specific configuration for a tool
func (l *GetTenantConfigLogic) GetTenantConfig(req *types.GetTenantConfigReq) (*types.TenantToolConfig, error) {
	l.Infof("GetTenantConfig called with name: %s", req.Name)

	// TODO: Implement getting tenant config from storage/service
	return &types.TenantToolConfig{
		ToolName: req.Name,
	}, nil
}

type SetTenantConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetTenantConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetTenantConfigLogic {
	return &SetTenantConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetTenantConfig sets tenant-specific configuration for a tool
func (l *SetTenantConfigLogic) SetTenantConfig(req *types.SetTenantConfigReq) (*types.SetTenantConfigResp, error) {
	l.Infof("SetTenantConfig called with name: %s", req.Name)

	// TODO: Implement setting tenant config in storage/service
	return &types.SetTenantConfigResp{
		Status: "configured",
	}, nil
}

type DeleteTenantConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTenantConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantConfigLogic {
	return &DeleteTenantConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteTenantConfig deletes tenant-specific configuration for a tool
func (l *DeleteTenantConfigLogic) DeleteTenantConfig(req *types.DeleteTenantConfigReq) (*types.DeleteTenantConfigResp, error) {
	l.Infof("DeleteTenantConfig called with name: %s", req.Name)

	// TODO: Implement deleting tenant config from storage/service
	return &types.DeleteTenantConfigResp{
		Status: "deleted",
	}, nil
}
