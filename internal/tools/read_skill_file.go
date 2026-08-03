package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/pomclaw/pomclaw/internal/contracts"
)

// ReadSkillFileTool reads a single file from a skill's package (OSS zip) by
// relative path, e.g. references or scripts. Used so the agent can consult a
// skill's bundled reference docs without those files ever touching disk.
type ReadSkillFileTool struct {
	loader contracts.SkillsLoaderInterface
}

func NewReadSkillFileTool(loader contracts.SkillsLoaderInterface) tool.InvokableTool {
	return &ReadSkillFileTool{loader: loader}
}

func (t *ReadSkillFileTool) Name() string { return "read_skill_file" }

func (t *ReadSkillFileTool) Description() string {
	return "Read a file bundled inside a skill package (from the available-skills list). " +
		"Use this to consult a skill's reference docs, e.g. image prompt rules. " +
		"file_path is relative to the skill root (e.g. \"references/image-rules.md\"), without the top-level directory prefix."
}

func (t *ReadSkillFileTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: t.Name(),
		Desc: t.Description(),
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"skill_name": {
				Type:     schema.String,
				Desc:     "The name of the skill (from the available-skills list in system prompt)",
				Required: true,
			},
			"file_path": {
				Type:     schema.String,
				Desc:     "Path of the file inside the skill, relative to the skill root (e.g. \"references/image-rules.md\")",
				Required: true,
			},
		}),
	}, nil
}

type readSkillFileInput struct {
	SkillName string `json:"skill_name"`
	FilePath  string `json:"file_path"`
}

func (t *ReadSkillFileTool) InvokableRun(_ context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	var input readSkillFileInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if input.SkillName == "" || input.FilePath == "" {
		return "", fmt.Errorf("skill_name and file_path are required")
	}

	data, found, err := t.loader.GetSkillFile(input.SkillName, input.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read skill file: %w", err)
	}
	if !found {
		return fmt.Sprintf("File '%s' not found in skill '%s'.", input.FilePath, input.SkillName), nil
	}

	return string(data), nil
}
