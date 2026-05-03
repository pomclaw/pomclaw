package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// DailyNoteWriter is the interface the write_daily_note tool needs.
type DailyNoteWriter interface {
	AppendToday(agentID string, content string) error
}

type WriteDailyNoteInput struct {
	Content string `json:"content"`
}

type WriteDailyNoteOutput struct {
	Message string `json:"message"`
}

func NewWriteDailyNoteTool(store DailyNoteWriter) tool.InvokableTool {
	return utils.WrapInvokableToolWithErrorHandler(utils.NewTool[WriteDailyNoteInput, WriteDailyNoteOutput](
		&schema.ToolInfo{
			Name: "write_daily_note",
			Desc: "Append a note to today's daily journal. Use this to record events, tasks completed, observations, or anything worth noting for today. Notes are stored persistently and included in future context.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"content": {
					Type:     schema.String,
					Desc:     "The note content to append to today's daily journal",
					Required: true,
				},
			}),
		},
		func(ctx context.Context, input WriteDailyNoteInput) (WriteDailyNoteOutput, error) {
			if input.Content == "" {
				return WriteDailyNoteOutput{}, fmt.Errorf("content parameter is required")
			}

			if err := store.AppendToday(AgentIDFromContext(ctx), input.Content); err != nil {
				return WriteDailyNoteOutput{}, fmt.Errorf("failed to write daily note: %w", err)
			}

			return WriteDailyNoteOutput{
				Message: fmt.Sprintf("Daily note written: %s", truncate(input.Content, 100)),
			}, nil
		},
	), func(ctx context.Context, err error) string { return err.Error() })
}
