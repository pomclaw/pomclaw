package toolsmanager

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/tools"
)

type ToolsManager struct {
	memoryStore     contracts.SqlMemoryStore
	toolGrantsModel model.ToolGrantsModel
	agentsModel     model.AgentsModel
}

func NewToolsManager(memoryStore contracts.SqlMemoryStore, toolGrantsModel model.ToolGrantsModel, agentsModel model.AgentsModel) contracts.ToolsManagerInterface {
	return &ToolsManager{
		memoryStore:     memoryStore,
		toolGrantsModel: toolGrantsModel,
		agentsModel:     agentsModel,
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
	const restrict = false

	// 有序工具列表
	toolDefs := []toolDef{
		{"read_file", true, func() tool.BaseTool { return tools.NewReadFileTool(restrict) }},
		{"write_file", true, func() tool.BaseTool { return tools.NewWriteFileTool(restrict) }},
		{"list_dir", true, func() tool.BaseTool { return tools.NewListDirTool(restrict) }},
		{"edit_file", true, func() tool.BaseTool { return tools.NewEditFileTool(restrict) }},
		{"append_file", true, func() tool.BaseTool { return tools.NewAppendFileTool(restrict) }},
		{"exec", true, func() tool.BaseTool { return tools.NewExecTool(restrict) }},
		{"remember", true, func() tool.BaseTool { return tools.NewRememberTool(&rememberAdapter{store: t.memoryStore}) }},
		{"write_daily_note", true, func() tool.BaseTool { return tools.NewWriteDailyNoteTool(t.memoryStore) }},
		{"recall", true, func() tool.BaseTool { return tools.NewRecallTool(&recallAdapter{store: t.memoryStore}) }},
	}

	// 如果没有 userId，从 agentID 查询
	if userId == "" {
		agent, _ := t.agentsModel.FindOne(ctx, agentID)
		if agent != nil {
			userId = agent.UserId
		}
	}

	// 一次性查询所有用户工具授权
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
	toolsNodeConfig.Tools = append(toolsNodeConfig.Tools, []tool.BaseTool{
		tools.NewReadFileTool(restrict),
		tools.NewWriteFileTool(restrict),
		tools.NewListDirTool(restrict),
		tools.NewEditFileTool(restrict),
		tools.NewAppendFileTool(restrict),
		tools.NewExecTool(restrict),
		tools.NewRememberTool(&rememberAdapter{store: t.memoryStore}),
		tools.NewWriteDailyNoteTool(t.memoryStore),
		tools.NewRecallTool(&recallAdapter{store: t.memoryStore}),
	}...)
	return toolsNodeConfig
}
