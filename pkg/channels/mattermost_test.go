package channels

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/config"
)

func TestParseMattermostChatID(t *testing.T) {
	tests := []struct {
		name       string
		chatID     string
		wantChanID string
		wantRootID string
	}{
		{
			name:       "channel only",
			chatID:     "f6c1msw84pdcjeuw4gomnaj6se",
			wantChanID: "f6c1msw84pdcjeuw4gomnaj6se",
			wantRootID: "",
		},
		{
			name:       "channel with thread",
			chatID:     "f6c1msw84pdcjeuw4gomnaj6se/abcdef1234567890",
			wantChanID: "f6c1msw84pdcjeuw4gomnaj6se",
			wantRootID: "abcdef1234567890",
		},
		{
			name:       "empty string",
			chatID:     "",
			wantChanID: "",
			wantRootID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chanID, rootID := parseMattermostChatID(tt.chatID)
			if chanID != tt.wantChanID {
				t.Errorf("parseMattermostChatID(%q) channelID = %q, want %q", tt.chatID, chanID, tt.wantChanID)
			}
			if rootID != tt.wantRootID {
				t.Errorf("parseMattermostChatID(%q) rootID = %q, want %q", tt.chatID, rootID, tt.wantRootID)
			}
		})
	}
}

func TestMattermostWSURL(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		wantWS    string
	}{
		{
			name:      "https to wss",
			serverURL: "https://mm.example.com",
			wantWS:    "wss://mm.example.com",
		},
		{
			name:      "http to ws",
			serverURL: "http://localhost:8065",
			wantWS:    "ws://localhost:8065",
		},
		{
			name:      "with path",
			serverURL: "https://mm.example.com/mattermost",
			wantWS:    "wss://mm.example.com/mattermost",
		},
		{
			name:      "already wss",
			serverURL: "wss://mm.example.com",
			wantWS:    "wss://mm.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mattermostWSURL(tt.serverURL)
			if got != tt.wantWS {
				t.Errorf("mattermostWSURL(%q) = %q, want %q", tt.serverURL, got, tt.wantWS)
			}
		})
	}
}

func TestNewMattermostChannel(t *testing.T) {
	msgBus := bus.NewMessageBus()

	t.Run("missing server_url", func(t *testing.T) {
		cfg := config.MattermostConfig{
			ServerURL: "",
			Token:     "some-token",
		}
		_, err := NewMattermostChannel(cfg, msgBus)
		if err == nil {
			t.Error("expected error for missing server_url, got nil")
		}
	})

	t.Run("missing token", func(t *testing.T) {
		cfg := config.MattermostConfig{
			ServerURL: "https://mm.example.com",
			Token:     "",
		}
		_, err := NewMattermostChannel(cfg, msgBus)
		if err == nil {
			t.Error("expected error for missing token, got nil")
		}
	})

	t.Run("valid config", func(t *testing.T) {
		cfg := config.MattermostConfig{
			ServerURL: "https://mm.example.com",
			Token:     "test-token",
			AllowFrom: []string{"user1"},
		}
		ch, err := NewMattermostChannel(cfg, msgBus)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ch.Name() != "mattermost" {
			t.Errorf("Name() = %q, want %q", ch.Name(), "mattermost")
		}
		if ch.IsRunning() {
			t.Error("new channel should not be running")
		}
	})
}

func TestMattermostChannelIsAllowed(t *testing.T) {
	msgBus := bus.NewMessageBus()

	t.Run("empty allowlist allows all", func(t *testing.T) {
		cfg := config.MattermostConfig{
			ServerURL: "https://mm.example.com",
			Token:     "test-token",
			AllowFrom: []string{},
		}
		ch, _ := NewMattermostChannel(cfg, msgBus)
		if !ch.IsAllowed("any-user-id") {
			t.Error("empty allowlist should allow all users")
		}
	})

	t.Run("allowlist restricts users", func(t *testing.T) {
		cfg := config.MattermostConfig{
			ServerURL: "https://mm.example.com",
			Token:     "test-token",
			AllowFrom: []string{"allowed-user"},
		}
		ch, _ := NewMattermostChannel(cfg, msgBus)
		if !ch.IsAllowed("allowed-user") {
			t.Error("allowed user should pass allowlist check")
		}
		if ch.IsAllowed("blocked-user") {
			t.Error("non-allowed user should be blocked")
		}
	})
}

func TestMattermostSendNotRunning(t *testing.T) {
	msgBus := bus.NewMessageBus()
	cfg := config.MattermostConfig{
		ServerURL: "https://mm.example.com",
		Token:     "test-token",
	}
	ch, _ := NewMattermostChannel(cfg, msgBus)

	err := ch.Send(context.Background(), bus.OutboundMessage{
		ChatID:  "some-channel-id",
		Content: "hello",
	})
	if err == nil {
		t.Error("expected error when sending on non-running channel")
	}
}

func TestMattermostSendInvalidChatID(t *testing.T) {
	msgBus := bus.NewMessageBus()
	cfg := config.MattermostConfig{
		ServerURL: "https://mm.example.com",
		Token:     "test-token",
	}
	ch, _ := NewMattermostChannel(cfg, msgBus)
	ch.setRunning(true)

	err := ch.Send(context.Background(), bus.OutboundMessage{
		ChatID:  "",
		Content: "hello",
	})
	if err == nil {
		t.Error("expected error for empty chat ID")
	}
}

// --- Integration tests below require env vars ---
// MM_SERVER_URL, MM_BOT_TOKEN, MM_CHANNEL_ID
// Optional: MM_USER_TOKEN (for E2E test)
//
// Run: MM_SERVER_URL=... MM_BOT_TOKEN=... MM_CHANNEL_ID=... go test ./pkg/channels/ -run TestMattermost -v

func mmEnv(t *testing.T) (serverURL, botToken, channelID string) {
	t.Helper()
	serverURL = os.Getenv("MM_SERVER_URL")
	botToken = os.Getenv("MM_BOT_TOKEN")
	channelID = os.Getenv("MM_CHANNEL_ID")
	if serverURL == "" || botToken == "" || channelID == "" {
		t.Skip("MM_SERVER_URL, MM_BOT_TOKEN, MM_CHANNEL_ID not set, skipping integration test")
	}
	return
}

func TestMattermostIntegration(t *testing.T) {
	serverURL, botToken, channelID := mmEnv(t)

	cfg := config.MattermostConfig{
		Enabled:   true,
		ServerURL: serverURL,
		Token:     botToken,
	}

	msgBus := bus.NewMessageBus()

	ch, err := NewMattermostChannel(cfg, msgBus)
	if err != nil {
		t.Fatalf("NewMattermostChannel failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("start and auth", func(t *testing.T) {
		if err := ch.Start(ctx); err != nil {
			t.Fatalf("Start failed: %v", err)
		}
		defer ch.Stop(ctx)

		if !ch.IsRunning() {
			t.Error("channel should be running after Start")
		}
		if ch.botUserID == "" {
			t.Error("botUserID should be set after Start")
		}
		t.Logf("bot_user_id: %s", ch.botUserID)
	})

	t.Run("send message", func(t *testing.T) {
		ch2, err := NewMattermostChannel(cfg, msgBus)
		if err != nil {
			t.Fatalf("NewMattermostChannel failed: %v", err)
		}
		if err := ch2.Start(ctx); err != nil {
			t.Fatalf("Start failed: %v", err)
		}
		defer ch2.Stop(ctx)

		if err := ch2.Send(ctx, bus.OutboundMessage{
			Channel: "mattermost",
			ChatID:  channelID,
			Content: "[test] Mattermost channel integration test",
		}); err != nil {
			t.Errorf("Send failed: %v", err)
		}
	})
}

// E2E: user sends message via MM_USER_TOKEN, bot receives it via WebSocket on bus
func TestMattermostE2E(t *testing.T) {
	serverURL, botToken, channelID := mmEnv(t)
	userToken := os.Getenv("MM_USER_TOKEN")
	if userToken == "" {
		t.Skip("MM_USER_TOKEN not set, skipping E2E test")
	}

	msgBus := bus.NewMessageBus()

	cfg := config.MattermostConfig{
		Enabled:   true,
		ServerURL: serverURL,
		Token:     botToken,
	}
	ch, err := NewMattermostChannel(cfg, msgBus)
	if err != nil {
		t.Fatalf("NewMattermostChannel failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := ch.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer ch.Stop(context.Background())

	time.Sleep(1 * time.Second)

	userClient := model.NewAPIv4Client(serverURL)
	userClient.SetToken(userToken)

	testMsg := fmt.Sprintf("[e2e-test] hello from user at %s", time.Now().Format(time.RFC3339))
	createdPost, _, err := userClient.CreatePost(ctx, &model.Post{
		ChannelId: channelID,
		Message:   testMsg,
	})
	if err != nil {
		t.Fatalf("User send message failed: %v", err)
	}
	t.Logf("User sent post_id=%s, message=%q", createdPost.Id, testMsg)

	recvCtx, recvCancel := context.WithTimeout(ctx, 10*time.Second)
	defer recvCancel()

	msg, ok := msgBus.ConsumeInbound(recvCtx)
	if !ok {
		t.Fatal("Timeout: bot did not receive the user's message within 10s")
	}

	t.Logf("Bot received: sender=%s, chatID=%s, content=%q", msg.SenderID, msg.ChatID, msg.Content)

	if msg.Channel != "mattermost" {
		t.Errorf("expected channel=mattermost, got %q", msg.Channel)
	}
	if msg.Content != testMsg {
		t.Errorf("content mismatch:\n  want: %q\n  got:  %q", testMsg, msg.Content)
	}
	if msg.Metadata["post_id"] != createdPost.Id {
		t.Errorf("post_id mismatch: want %q, got %q", createdPost.Id, msg.Metadata["post_id"])
	}
}

// Long-lived listener: bot keeps WebSocket open, receives messages from the web UI and replies.
// Run with: MM_SERVER_URL=... MM_BOT_TOKEN=... MM_CHANNEL_ID=... go test ./pkg/channels/ -run TestMattermostListen -v -timeout 300s
func TestMattermostListen(t *testing.T) {
	serverURL, botToken, _ := mmEnv(t)
	listenDuration := 100 * time.Second

	msgBus := bus.NewMessageBus()

	cfg := config.MattermostConfig{
		Enabled:   true,
		ServerURL: serverURL,
		Token:     botToken,
	}
	ch, err := NewMattermostChannel(cfg, msgBus)
	if err != nil {
		t.Fatalf("NewMattermostChannel failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), listenDuration)
	defer cancel()

	if err := ch.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer ch.Stop(context.Background())

	t.Logf("Bot listening (bot_user_id=%s). Send messages from the Mattermost web page now...", ch.botUserID)
	t.Logf("Will print received messages for %s. Ctrl+C to stop early.", listenDuration)

	count := 0
	for {
		msg, ok := msgBus.ConsumeInbound(ctx)
		if !ok {
			break
		}
		count++
		t.Logf("[#%d] sender=%s chatID=%s content=%q metadata=%v",
			count, msg.SenderID, msg.ChatID, msg.Content, msg.Metadata)

		reply := fmt.Sprintf("我收到了「%s」的消息", msg.Content)
		if err := ch.Send(ctx, bus.OutboundMessage{
			Channel: "mattermost",
			ChatID:  msg.ChatID,
			Content: reply,
		}); err != nil {
			t.Errorf("Bot reply failed: %v", err)
		} else {
			t.Logf("[#%d] Bot replied: %s", count, reply)
		}
	}

	t.Logf("Done. Received %d message(s) total.", count)
}
