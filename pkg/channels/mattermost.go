package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/channels/base"
	"github.com/pomclaw/pomclaw/pkg/config"
	"github.com/pomclaw/pomclaw/pkg/logger"
	"github.com/pomclaw/pomclaw/pkg/utils"
)

const maxFileSize = 50 * 1024 * 1024 // 50MB

type MattermostChannel struct {
	*base.BaseChannel
	config      config.MattermostConfig
	client      *model.Client4
	mu          sync.RWMutex
	wsClient    *model.WebSocketClient
	botUserID   string
	ctx         context.Context
	cancel      context.CancelFunc
	loopDone    chan struct{} // closed when eventLoop exits
	pendingAcks sync.Map
}

type mattermostMessageRef struct {
	PostID string
}

func NewMattermostChannel(cfg config.MattermostConfig, messageBus *bus.MessageBus) (*MattermostChannel, error) {
	if cfg.ServerURL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("mattermost server_url and token are required")
	}

	client := model.NewAPIv4Client(cfg.ServerURL)
	client.SetToken(cfg.Token)

	baseChannel := base.NewBaseChannel("mattermost", cfg, messageBus, cfg.AllowFrom)

	return &MattermostChannel{
		BaseChannel: baseChannel,
		config:      cfg,
		client:      client,
	}, nil
}

func (c *MattermostChannel) Start(ctx context.Context) error {
	if c.IsRunning() {
		return fmt.Errorf("mattermost channel already running")
	}

	logger.InfoC("mattermost", "Starting Mattermost channel")

	childCtx, cancel := context.WithCancel(ctx)

	user, _, err := c.client.GetMe(childCtx, "")
	if err != nil {
		cancel()
		return fmt.Errorf("mattermost auth failed: %w", err)
	}
	c.botUserID = user.Id
	c.ctx, c.cancel = childCtx, cancel

	logger.InfoCF("mattermost", "Mattermost bot connected", map[string]interface{}{
		"bot_user_id": c.botUserID,
		"username":    user.Username,
	})

	wsURL := mattermostWSURL(c.config.ServerURL)
	wsClient, err := model.NewWebSocketClient4(wsURL, c.config.Token)
	if err != nil {
		c.cancel()
		c.ctx, c.cancel = nil, nil
		return fmt.Errorf("mattermost websocket connection failed: %w", err)
	}

	c.mu.Lock()
	c.wsClient = wsClient
	c.mu.Unlock()

	wsClient.Listen()

	c.loopDone = make(chan struct{})
	go c.eventLoop()

	c.SetRunning(true)
	logger.InfoC("mattermost", "Mattermost channel started")
	return nil
}

func (c *MattermostChannel) Stop(_ context.Context) error {
	logger.InfoC("mattermost", "Stopping Mattermost channel")

	c.SetRunning(false)

	if c.cancel != nil {
		c.cancel()
	}

	c.mu.Lock()
	ws := c.wsClient
	c.wsClient = nil
	c.mu.Unlock()

	closeWS(ws)

	// Wait for eventLoop goroutine to fully exit so a subsequent Start()
	// cannot race on c.ctx / c.wsClient with a still-running old loop.
	if c.loopDone != nil {
		<-c.loopDone
	}

	logger.InfoC("mattermost", "Mattermost channel stopped")
	return nil
}

// closeWS forces a non-blocking close on a WebSocket client.
func closeWS(ws *model.WebSocketClient) {
	if ws == nil {
		return
	}
	if ws.Conn != nil {
		_ = ws.Conn.SetReadDeadline(time.Now())
	}
	ws.Close()
}

func (c *MattermostChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() {
		return fmt.Errorf("mattermost channel not running")
	}

	channelID, rootID := parseMattermostChatID(msg.ChatID)
	if channelID == "" {
		return fmt.Errorf("invalid mattermost chat ID: %s", msg.ChatID)
	}

	post := &model.Post{
		ChannelId: channelID,
		Message:   msg.Content,
	}
	if rootID != "" {
		post.RootId = rootID
	}

	_, _, err := c.client.CreatePost(ctx, post)
	if err != nil {
		return fmt.Errorf("failed to send mattermost message: %w", err)
	}

	if ref, ok := c.pendingAcks.LoadAndDelete(msg.ChatID); ok {
		msgRef := ref.(mattermostMessageRef)
		if _, _, err := c.client.SaveReaction(ctx, &model.Reaction{
			UserId:    c.botUserID,
			PostId:    msgRef.PostID,
			EmojiName: "white_check_mark",
		}); err != nil {
			logger.DebugCF("mattermost", "Failed to add check mark reaction", map[string]interface{}{
				"post_id": msgRef.PostID,
				"error":   err.Error(),
			})
		}
	}

	logger.DebugCF("mattermost", "Message sent", map[string]interface{}{
		"channel_id": channelID,
		"root_id":    rootID,
	})

	return nil
}

func (c *MattermostChannel) eventLoop() {
	defer close(c.loopDone)
	for {
		// Snapshot wsClient under lock so select reads consistent channels.
		c.mu.RLock()
		ws := c.wsClient
		c.mu.RUnlock()

		if ws == nil {
			return
		}

		select {
		case <-c.ctx.Done():
			return
		case event, ok := <-ws.EventChannel:
			if !ok {
				if c.reconnect() {
					continue
				}
				return
			}
			if event == nil {
				continue
			}
			// Filter bot's own messages early to avoid echo loops.
			if event.EventType() == model.WebsocketEventPosted {
				c.handlePosted(event)
			}
		case <-ws.PingTimeoutChannel:
			logger.WarnC("mattermost", "WebSocket ping timeout, reconnecting")
			if c.reconnect() {
				continue
			}
			return
		}
	}
}

// reconnect replaces the WebSocket connection in-place and returns true on success.
// Called from eventLoop goroutine only — the mutex protects against concurrent Stop.
func (c *MattermostChannel) reconnect() bool {
	if c.ctx.Err() != nil {
		return false
	}

	// Close old connection.
	c.mu.Lock()
	old := c.wsClient
	c.wsClient = nil
	c.mu.Unlock()

	closeWS(old)

	for i := 0; ; i++ {
		delay := time.NewTimer(time.Duration(min(i+1, 30)) * time.Second)
		select {
		case <-c.ctx.Done():
			delay.Stop()
			return false
		case <-delay.C:
		}

		// Re-check after waking: Stop() may have fired concurrently with the timer.
		if c.ctx.Err() != nil {
			return false
		}

		logger.InfoCF("mattermost", "Attempting WebSocket reconnect", map[string]interface{}{
			"attempt": i + 1,
		})

		wsURL := mattermostWSURL(c.config.ServerURL)
		wsClient, err := model.NewWebSocketClient4(wsURL, c.config.Token)
		if err != nil {
			logger.ErrorCF("mattermost", "WebSocket reconnect failed", map[string]interface{}{
				"error": err.Error(),
			})
			continue
		}

		// Hold the lock while checking ctx to establish happens-before with Stop().
		// Stop() calls cancel() then grabs mu to nil-out wsClient.  If ctx is
		// already cancelled here, Stop() either already ran (wsClient would leak)
		// or is about to run but hasn't grabbed mu yet — either way we must not
		// store the new connection.
		c.mu.Lock()
		if c.ctx.Err() != nil {
			c.mu.Unlock()
			closeWS(wsClient)
			return false
		}
		c.wsClient = wsClient
		c.mu.Unlock()

		wsClient.Listen()
		logger.InfoC("mattermost", "WebSocket reconnected successfully")
		return true
	}
}

func (c *MattermostChannel) handlePosted(event *model.WebSocketEvent) {
	rawPost, ok := event.GetData()["post"].(string)
	if !ok {
		return
	}

	var post model.Post
	if err := json.Unmarshal([]byte(rawPost), &post); err != nil {
		logger.ErrorCF("mattermost", "Failed to parse post", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Skip bot's own messages to prevent echo loops.
	if post.UserId == c.botUserID {
		return
	}

	if !c.IsAllowed(post.UserId) {
		logger.DebugCF("mattermost", "Message rejected by allowlist", map[string]interface{}{
			"user_id": post.UserId,
		})
		return
	}

	channelID := post.ChannelId
	rootID := post.RootId
	chatID := channelID
	if rootID != "" {
		chatID = channelID + "/" + rootID
	}

	if _, _, err := c.client.SaveReaction(c.ctx, &model.Reaction{
		UserId:    c.botUserID,
		PostId:    post.Id,
		EmojiName: "eyes",
	}); err != nil {
		logger.DebugCF("mattermost", "Failed to add eyes reaction", map[string]interface{}{
			"post_id": post.Id,
			"error":   err.Error(),
		})
	}

	c.pendingAcks.Store(chatID, mattermostMessageRef{PostID: post.Id})

	content := post.Message
	var mediaPaths []string
	var localFiles []string

	defer func() {
		for _, file := range localFiles {
			if err := os.Remove(file); err != nil {
				logger.DebugCF("mattermost", "Failed to cleanup temp file", map[string]interface{}{
					"file":  file,
					"error": err.Error(),
				})
			}
		}
	}()

	if len(post.FileIds) > 0 {
		fileInfos, _, err := c.client.GetFileInfosForPost(c.ctx, post.Id, "")
		if err != nil {
			logger.ErrorCF("mattermost", "Failed to get file infos", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			for _, fi := range fileInfos {
				if fi.Size > maxFileSize {
					logger.WarnCF("mattermost", "File too large, skipping", map[string]interface{}{
						"file":     fi.Name,
						"size":     fi.Size,
						"max_size": maxFileSize,
					})
					content += fmt.Sprintf("\n[file: %s (too large, skipped)]", fi.Name)
					continue
				}
				localPath := c.downloadFile(fi.Id, fi.Name)
				if localPath == "" {
					continue
				}
				localFiles = append(localFiles, localPath)
				mediaPaths = append(mediaPaths, localPath)

				if utils.IsAudioFile(fi.Name, fi.MimeType) {
					content += fmt.Sprintf("\n[audio: %s]", fi.Name)
				} else {
					content += fmt.Sprintf("\n[file: %s]", fi.Name)
				}
			}
		}
	}

	if strings.TrimSpace(content) == "" {
		return
	}

	metadata := map[string]string{
		"post_id":    post.Id,
		"user_id":    post.UserId,
		"channel_id": channelID,
		"root_id":    rootID,
		"platform":   "mattermost",
	}

	logger.DebugCF("mattermost", "Received message", map[string]interface{}{
		"sender_id":  post.UserId,
		"chat_id":    chatID,
		"preview":    utils.Truncate(content, 50),
		"has_thread": rootID != "",
	})

	c.HandleMessage(post.UserId, chatID, content, mediaPaths, metadata)
}

func (c *MattermostChannel) downloadFile(fileID, filename string) string {
	data, _, err := c.client.GetFile(c.ctx, fileID)
	if err != nil {
		logger.ErrorCF("mattermost", "Failed to download file", map[string]interface{}{
			"file_id": fileID,
			"error":   err.Error(),
		})
		return ""
	}

	// Sanitize filename to prevent path traversal.
	safeName := filepath.Base(filename)

	tmpFile, err := os.CreateTemp("", "mm-*-"+safeName)
	if err != nil {
		logger.ErrorCF("mattermost", "Failed to create temp file", map[string]interface{}{
			"error": err.Error(),
		})
		return ""
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		logger.ErrorCF("mattermost", "Failed to write temp file", map[string]interface{}{
			"error": err.Error(),
		})
		os.Remove(tmpFile.Name())
		return ""
	}

	return tmpFile.Name()
}

func parseMattermostChatID(chatID string) (channelID, rootID string) {
	parts := strings.SplitN(chatID, "/", 2)
	channelID = parts[0]
	if len(parts) > 1 {
		rootID = parts[1]
	}
	return
}

func mattermostWSURL(serverURL string) string {
	switch {
	case strings.HasPrefix(serverURL, "https://"):
		return "wss://" + serverURL[len("https://"):]
	case strings.HasPrefix(serverURL, "http://"):
		return "ws://" + serverURL[len("http://"):]
	default:
		return serverURL
	}
}
