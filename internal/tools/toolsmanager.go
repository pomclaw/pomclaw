package tools

import (
	"context"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/pomclaw/pomclaw/internal/contracts"
	"github.com/pomclaw/pomclaw/internal/model"
)

type ToolsManager struct {
	toolGrantsModel        model.ToolGrantsModel
	agentsModel            model.AgentsModel
	memoryStore            contracts.SqlMemoryStore
	memoryDocumentsModel   model.MemoryDocumentsModel
	agentContextFilesModel model.AgentContextFilesModel
}

func NewToolsManager(toolGrantsModel model.ToolGrantsModel, agentsModel model.AgentsModel, memoryStore contracts.SqlMemoryStore, memoryDocumentsModel model.MemoryDocumentsModel, agentContextFilesModel model.AgentContextFilesModel) contracts.ToolsManagerInterface {
	return &ToolsManager{
		toolGrantsModel:        toolGrantsModel,
		agentsModel:            agentsModel,
		memoryStore:            memoryStore,
		memoryDocumentsModel:   memoryDocumentsModel,
		agentContextFilesModel: agentContextFilesModel,
	}
}

type toolDef struct {
	name    string
	enabled bool
	builder func() tool.BaseTool
}

func (t ToolsManager) GetToolsToolDef(ctx context.Context, userId, agentID string) []contracts.ToolDef {
	toolDefs := t.tools(ctx, userId, agentID)

	var result []contracts.ToolDef
	for _, def := range toolDefs {
		tt := def.builder()
		info, err := tt.Info(ctx)
		if err != nil {
			continue
		}
		result = append(result, contracts.ToolDef{
			Name:    info.Name,
			Display: info.Name,
			Desc:    info.Desc,
			Enabled: def.enabled,
		})
	}
	return result
}

func (t ToolsManager) GetTools(ctx context.Context, userId, agentID string) compose.ToolsNodeConfig {
	toolDefs := t.tools(ctx, userId, agentID)
	toolsNodeConfig := compose.ToolsNodeConfig{}
	for _, def := range toolDefs {
		if def.enabled {
			toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, def.builder())
		}
	}
	return toolsNodeConfig
}

func (t ToolsManager) tools(ctx context.Context, userId, agentID string) []toolDef {
	// 从全量工具列表转换为 toolDef（默认全部启用）
	var toolDefs []toolDef
	for _, tt := range t.buildDefaultTools(false).Tools {
		tt := tt
		info, err := tt.Info(ctx)
		if err != nil {
			continue
		}
		toolDefs = append(toolDefs, toolDef{
			name:    info.Name,
			enabled: true,
			builder: func() tool.BaseTool { return tt },
		})
	}

	// 如果没有 userId，从 agentID 查询
	if userId == "" {
		agent, _ := t.agentsModel.FindOne(ctx, agentID)
		if agent != nil {
			userId = agent.UserId
		}
	}

	// 一次性查询所有用户工具授权并应用过滤
	grants, _ := t.toolGrantsModel.FindAllByUserId(ctx, userId)
	for _, grant := range grants {
		if grant.Enabled.Valid && !grant.Enabled.Bool {
			for i := range toolDefs {
				if toolDefs[i].name == grant.ToolName {
					toolDefs[i].enabled = false
				}
			}
		}
	}

	return toolDefs
}

func (t ToolsManager) buildDefaultTools(restrict bool) compose.ToolsNodeConfig {
	toolsNodeConfig := compose.ToolsNodeConfig{}

	// Create interceptors for virtual filesystem routing
	contextFileIntc := NewContextFileInterceptor(t.agentContextFilesModel)

	// 完整的工具列表（无权限过滤，作为基础）
	toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, []tool.BaseTool{
		// contextFile
		utils.WrapInvokableToolWithErrorHandler(NewReadFileTool(restrict, contextFileIntc), errorHandler),
		utils.WrapInvokableToolWithErrorHandler(NewWriteFileTool(restrict, contextFileIntc), errorHandler),
		utils.WrapInvokableToolWithErrorHandler(NewListFilesTool(restrict, contextFileIntc), errorHandler),
		utils.WrapInvokableToolWithErrorHandler(NewEditTool(restrict, contextFileIntc), errorHandler),
		// memory
		utils.WrapInvokableToolWithErrorHandler(NewRememberTool(&rememberAdapter{store: t.memoryStore}), errorHandler),
		utils.WrapInvokableToolWithErrorHandler(NewRecallTool(&recallAdapter{store: t.memoryStore}), errorHandler),
		// exec
		utils.WrapInvokableToolWithErrorHandler(NewExecTool(restrict), errorHandler),
	}...)
	return toolsNodeConfig
}

var errorHandler = func(ctx context.Context, err error) string { return err.Error() }
