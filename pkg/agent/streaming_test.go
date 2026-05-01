package agent

import (
	"context"
	"os"
	"testing"

	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/bus"
)

type streamingTestProvider struct {
	chatCalled       bool
	chatStreamCalled bool
	chunks           []string
	response         *providers.LLMResponse
}

func (p *streamingTestProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	p.chatCalled = true
	if p.response != nil {
		return p.response, nil
	}
	return &providers.LLMResponse{Content: "fallback"}, nil
}

func (p *streamingTestProvider) ChatStream(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}, onChunk func(accumulated string)) (*providers.LLMResponse, error) {
	p.chatStreamCalled = true
	for _, chunk := range p.chunks {
		onChunk(chunk)
	}
	if p.response != nil {
		return p.response, nil
	}
	return &providers.LLMResponse{Content: "streamed"}, nil
}

func (p *streamingTestProvider) GetDefaultModel() string {
	return "test-model"
}

type recordingStreamer struct {
	updates   []string
	finalized []string
	canceled  bool
}

func (s *recordingStreamer) Update(_ context.Context, content string) error {
	s.updates = append(s.updates, content)
	return nil
}

func (s *recordingStreamer) Finalize(_ context.Context, content string) error {
	s.finalized = append(s.finalized, content)
	return nil
}

func (s *recordingStreamer) Cancel(_ context.Context) {
	s.canceled = true
}

func (s *recordingStreamer) HasPosted() bool {
	return len(s.updates) > 0
}

type recordingStreamDelegate struct {
	streamer *recordingStreamer
	ok       bool
}

func (d *recordingStreamDelegate) GetStreamer(_ context.Context, channel, chatID string) (bus.Streamer, bool) {
	if !d.ok || channel == "" || chatID == "" {
		return nil, false
	}
	return d.streamer, true
}

type noopSkillsLoader struct{}

func (noopSkillsLoader) ListSkills(string) []SkillInfo                { return nil }
func (noopSkillsLoader) LoadSkill(string, string) (string, bool)      { return "", false }
func (noopSkillsLoader) LoadSkillsForContext(string, []string) string { return "" }
func (noopSkillsLoader) BuildSkillsSummary(string) string             { return "" }

func newStreamingTestLoop(t *testing.T, provider *streamingTestProvider, streamer *recordingStreamer, delegateOK bool) (*AgentLoop, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "agent-streaming-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:         tmpDir,
				Model:             "test-model",
				MaxTokens:         4096,
				MaxToolIterations: 10,
			},
		},
	}

	msgBus := bus.NewMessageBus()
	msgBus.SetStreamDelegate(&recordingStreamDelegate{
		streamer: streamer,
		ok:       delegateOK,
	})
	al := NewAgentLoop(cfg, nil, msgBus, provider)
	al.SetSkillsLoader(noopSkillsLoader{})
	return al, func() { _ = os.RemoveAll(tmpDir) }
}

func TestProcessMessage_UsesStreamingProviderWhenStreamerExists(t *testing.T) {
	provider := &streamingTestProvider{
		chunks:   []string{"Hel", "Hello"},
		response: &providers.LLMResponse{Content: "Hello"},
	}
	streamer := &recordingStreamer{}
	al, cleanup := newStreamingTestLoop(t, provider, streamer, true)
	defer cleanup()

	response, err := al.processMessage(context.Background(), bus.InboundMessage{
		Channel:    "mattermost",
		SenderID:   "user-1",
		ChatID:     "mattermost:test-session",
		Content:    "hello",
		SessionKey: "session-1",
	})
	if err != nil {
		t.Fatalf("processMessage() error = %v", err)
	}
	if response != "Hello" {
		t.Fatalf("response = %q, want %q", response, "Hello")
	}
	if !provider.chatStreamCalled {
		t.Fatal("expected ChatStream to be used")
	}
	if provider.chatCalled {
		t.Fatal("expected Chat() not to be used")
	}
	if len(streamer.updates) != 2 {
		t.Fatalf("stream updates = %#v, want 2 accumulated updates", streamer.updates)
	}
	if len(streamer.finalized) != 1 || streamer.finalized[0] != "Hello" {
		t.Fatalf("finalized = %#v, want final Hello", streamer.finalized)
	}
	if streamer.canceled {
		t.Fatal("streamer should not be canceled on direct response")
	}
}

func TestProcessMessage_FallbackToChatWithoutStreamer(t *testing.T) {
	provider := &streamingTestProvider{
		response: &providers.LLMResponse{Content: "No stream"},
	}
	streamer := &recordingStreamer{}
	al, cleanup := newStreamingTestLoop(t, provider, streamer, false)
	defer cleanup()

	response, err := al.processMessage(context.Background(), bus.InboundMessage{
		Channel:    "mattermost",
		SenderID:   "user-1",
		ChatID:     "mattermost:test-session",
		Content:    "hello",
		SessionKey: "session-2",
	})
	if err != nil {
		t.Fatalf("processMessage() error = %v", err)
	}
	if response != "No stream" {
		t.Fatalf("response = %q, want %q", response, "No stream")
	}
	if provider.chatStreamCalled {
		t.Fatal("expected ChatStream not to be used when streamer is unavailable")
	}
	if !provider.chatCalled {
		t.Fatal("expected Chat() to be used when streamer is unavailable")
	}
}

func TestProcessMessage_CancelStreamingWhenToolCallsFollow(t *testing.T) {
	provider := &streamingTestProvider{
		chunks: []string{"Thinking"},
		response: &providers.LLMResponse{
			Content: "Thinking",
			ToolCalls: []providers.ToolCall{
				{
					ID:        "call-1",
					Name:      "missing_tool",
					Arguments: map[string]interface{}{"q": "stream"},
				},
			},
		},
	}
	streamer := &recordingStreamer{}
	al, cleanup := newStreamingTestLoop(t, provider, streamer, true)
	defer cleanup()

	response, err := al.processMessage(context.Background(), bus.InboundMessage{
		Channel:    "mattermost",
		SenderID:   "user-1",
		ChatID:     "mattermost:test-session",
		Content:    "hello",
		SessionKey: "session-3",
	})
	if err != nil {
		t.Fatalf("processMessage() error = %v", err)
	}
	if response == "" {
		t.Fatal("expected follow-up tool result content after tool-call path")
	}
	if !provider.chatStreamCalled {
		t.Fatal("expected ChatStream to be used")
	}
	if !streamer.canceled {
		t.Fatal("expected streamer to be canceled when tool calls follow streamed text")
	}
	if len(streamer.finalized) != 0 {
		t.Fatalf("finalized = %#v, want no finalize after tool calls", streamer.finalized)
	}
}
