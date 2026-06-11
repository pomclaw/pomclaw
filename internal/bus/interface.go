package bus

import (
	"context"
	"github.com/cloudwego/eino/schema"
	"time"
)

// ==================== Publisher Interfaces ====================

// Streamer publishes outbound messages to channels
type Streamer interface {
	PublishRunStarted(ctx context.Context, payload *RunStartedPayload) error
	PublishRunCompleted(ctx context.Context, payload *RunCompletedPayload) error
	PublishToolCall(ctx context.Context, payload *ToolCallPayload) error
	PublishToolResult(ctx context.Context, payload *ToolResultPayload) error
	PublishChunk(ctx context.Context, payload *ChunkPayload) error
}

// "type": "run.started",
// "payload": {
// "message": "你仔细看下，"
// }

type RunStartedPayload struct {
	Message string `json:"message"`
}

//"type": "tool.call",
//
//"payload": {
//"arguments": {
//"command": "ls -laA"
//},
//"id": "call_5774133a46e48bcace758ddd31255386bb6",
//"name": "Bash"
//}

type ToolCallPayload struct {
	Arguments interface{} `json:"arguments"`
	Id        string      `json:"id"`
	Name      string      `json:"name"`
}

//"type": "tool.result",
//
//"payload": {
//"arguments": {
//"command": "ls -laA"
//},
//"id": "call_5774133a46e48bcace758ddd31255386bb6",
//"is_error": false,
//"name": "Bash",
//"result": "total 0\n"
//}

type ToolResultPayload struct {
	Arguments interface{} `json:"arguments"`
	Id        string      `json:"id"`
	IsError   bool        `json:"is_error"`
	Name      string      `json:"name"`
	Result    string      `json:"result"`
}

//"type": "chunk",
//
//"payload": {
//"content": "主人"
//}

type ChunkPayload struct {
	Content string `json:"content"`
}

//"type": "run.completed",
//
//"payload": {
//"content": "主人，小狐真的很仔细地看了好多遍呢～ 🦊💦\n\n用了多种方式确认：\n- `ls -la` → 空目录\n- `dir /a` → 没有文件\n- `tree /F` → 没有文件\n- `attrib` → 找不到任何文件\n- `dir /s /b` → 完全没有输出\n\n工作目录 `C:\\Users\\Administrator\\system` 确确实实是**空的**，连隐藏文件都没有哦。\n\n主人是不是记错目录了？还是有其他地方想让我看看？ 😏✨",
//"usage": {
//"cache_creation_tokens": 0,
//"cache_read_tokens": 0,
//"completion_tokens": 304,
//"prompt_tokens": 116675,
//"total_tokens": 116979
//}
//}

type RunCompletedPayload struct {
	Content string `json:"content"`
	Usage   Usage  `json:"usage"`
}

type Usage struct {
	CacheCreationTokens int `json:"cache_creation_tokens"`
	CacheReadTokens     int `json:"cache_read_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	PromptTokens        int `json:"prompt_tokens"`
	TotalTokens         int `json:"total_tokens"`
}

// Message  chat.history
type Message struct {
	Role       schema.RoleType   `json:"role"`
	Content    string            `json:"content"`
	ToolCalls  []ToolCallPayload `json:"tool_calls"`
	ToolCallId string            `json:"tool_call_id"` // tool_result 使用
	CreatedAt  time.Time         `json:"created_at"`
}

func ConvertMessages(history []schema.Message) []Message {
	result := make([]Message, len(history))
	for i, msg := range history {
		result[i] = Message{
			Role:       msg.Role,
			Content:    msg.Content,
			ToolCallId: msg.ToolCallID,
			CreatedAt:  time.Now(),
		}

		// Convert ToolCalls from schema.ToolCall to ToolCallPayload
		if len(msg.ToolCalls) > 0 {
			result[i].ToolCalls = make([]ToolCallPayload, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				result[i].ToolCalls[j] = ToolCallPayload{
					Id:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				}
			}
		}
	}
	return result
}
