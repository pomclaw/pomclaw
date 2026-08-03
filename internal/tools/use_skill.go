package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/contracts"
)

type UseSkillTool struct {
	loader contracts.SkillsLoaderInterface
}

func NewUseSkillTool(loader contracts.SkillsLoaderInterface) tool.InvokableTool {
	return &UseSkillTool{loader: loader}
}

func (t *UseSkillTool) Name() string { return "use_skill" }

func (t *UseSkillTool) Description() string {
	return "Load a skill by name. Returns the skill's full instructions for you to follow."
}

func (t *UseSkillTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"name": {
				Type:     schema.String,
				Desc:     "The name of the skill to load (from the available-skills list in system prompt)",
				Required: true,
			},
		}),
	}, nil
}

type useSkillInput struct {
	Name string `json:"name"`
}

func (t *UseSkillTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var input useSkillInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.Name == "" {
		return "", fmt.Errorf("name parameter is required")
	}

	content, found := t.loader.LoadSkill(input.Name)
	if !found {
		return fmt.Sprintf("Skill '%s' not found. Check available skills in system prompt.", input.Name), nil
	}

	return content, nil
}
