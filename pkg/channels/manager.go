// Pomclaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package channels

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"sync"
	"time"

	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/pkg/bus"
	"github.com/pomclaw/pomclaw/pkg/channels/base"
	"github.com/pomclaw/pomclaw/pkg/constants"
)

type Manager struct {
	channels     map[string]base.Channel
	bus          *bus.MessageBus
	config       *config.Config
	dispatchTask *asyncTask
	mu           sync.RWMutex
}

type asyncTask struct {
	cancel context.CancelFunc
}

func NewManager(cfg *config.Config, messageBus *bus.MessageBus) (*Manager, error) {
	m := &Manager{
		channels: make(map[string]base.Channel),
		bus:      messageBus,
		config:   cfg,
	}

	if err := m.initChannels(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) initChannels() error {
	logx.Info("channels", "Initializing channel manager")

	// Data-driven channel initialization to reduce code duplication
	type channelEntry struct {
		name    string
		enabled bool
		factory func() (base.Channel, error)
	}

	entries := []channelEntry{

		{"whatsapp", m.config.Channels.WhatsApp.Enabled && m.config.Channels.WhatsApp.BridgeURL != "",
			func() (base.Channel, error) { return NewWhatsAppChannel(m.config.Channels.WhatsApp, m.bus) }},
		{"feishu", m.config.Channels.Feishu.Enabled,
			func() (base.Channel, error) { return NewFeishuChannel(m.config.Channels.Feishu, m.bus) }},
		{"maixcam", m.config.Channels.MaixCam.Enabled,
			func() (base.Channel, error) { return NewMaixCamChannel(m.config.Channels.MaixCam, m.bus) }},
		{"qq", m.config.Channels.QQ.Enabled,
			func() (base.Channel, error) { return NewQQChannel(m.config.Channels.QQ, m.bus) }},
		{"dingtalk", m.config.Channels.DingTalk.Enabled && m.config.Channels.DingTalk.ClientID != "",
			func() (base.Channel, error) { return NewDingTalkChannel(m.config.Channels.DingTalk, m.bus) }},
		{"line", m.config.Channels.LINE.Enabled && m.config.Channels.LINE.ChannelAccessToken != "",
			func() (base.Channel, error) { return NewLINEChannel(m.config.Channels.LINE, m.bus) }},
		{"onebot", m.config.Channels.OneBot.Enabled && m.config.Channels.OneBot.WSUrl != "",
			func() (base.Channel, error) { return NewOneBotChannel(m.config.Channels.OneBot, m.bus) }},
		{"mattermost", m.config.Channels.Mattermost.Enabled && m.config.Channels.Mattermost.Token != "",
			func() (base.Channel, error) { return NewMattermostChannel(m.config.Channels.Mattermost, m.bus) }},
	}

	for _, entry := range entries {
		if !entry.enabled {
			continue
		}
		logx.Debug("channels", fmt.Sprintf("Attempting to initialize %s channel", entry.name), nil)
		ch, err := entry.factory()
		if err != nil {
			logx.Error("channels", fmt.Sprintf("Failed to initialize %s channel", entry.name),
				map[string]interface{}{"error": err.Error()})
		} else {
			m.channels[entry.name] = ch
			logx.Info("channels", fmt.Sprintf("%s channel enabled successfully", entry.name), nil)
		}
	}

	logx.Info("channels", "Channel initialization completed", map[string]interface{}{
		"enabled_channels": len(m.channels),
	})

	return nil
}

func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.channels) == 0 {
		logx.Info("channels", "No channels enabled")
		return
	}

	logx.Info("channels", "Starting all channels")

	dispatchCtx, cancel := context.WithCancel(context.Background())
	m.dispatchTask = &asyncTask{cancel: cancel}

	go m.dispatchOutbound(dispatchCtx)

	for name, channel := range m.channels {
		logx.Info("channels", "Starting channel", map[string]interface{}{
			"channel": name,
		})
		if err := channel.Start(context.Background()); err != nil {
			logx.Error("channels", "Failed to start channel", map[string]interface{}{
				"channel": name,
				"error":   err.Error(),
			})
		}
	}

	logx.Info("channels", "All channels started")

}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	logx.Info("channels", "Stopping all channels")

	if m.dispatchTask != nil {
		m.dispatchTask.cancel()
		m.dispatchTask = nil
	}

	for name, channel := range m.channels {
		logx.Info("channels", "Stopping channel", map[string]interface{}{
			"channel": name,
		})
		if err := channel.Stop(context.Background()); err != nil {
			logx.Error("channels", "Error stopping channel", map[string]interface{}{
				"channel": name,
				"error":   err.Error(),
			})
		}
	}

	logx.Info("channels", "All channels stopped")
}

func (m *Manager) dispatchOutbound(ctx context.Context) {
	logx.Info("channels", "Outbound dispatcher started")

	for {
		select {
		case <-ctx.Done():
			logx.Info("channels", "Outbound dispatcher stopped")
			return
		default:
			msg, ok := m.bus.SubscribeOutbound(ctx)
			if !ok {
				continue
			}

			// Silently skip internal channels
			if constants.IsInternalChannel(msg.Channel) {
				continue
			}

			m.mu.RLock()
			channel, exists := m.channels[msg.Channel]
			m.mu.RUnlock()

			if !exists {
				logx.Info("channels", "Unknown channel for outbound message", map[string]interface{}{
					"channel": msg.Channel,
				})
				continue
			}

			// Send with a 30-second timeout to prevent one slow channel from blocking dispatch
			sendCtx, sendCancel := context.WithTimeout(ctx, 30*time.Second)
			if err := channel.Send(sendCtx, msg); err != nil {
				logx.Error("channels", "Error sending message to channel", map[string]interface{}{
					"channel": msg.Channel,
					"error":   err.Error(),
				})
			}
			sendCancel()
		}
	}
}

func (m *Manager) GetChannel(name string) (base.Channel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	channel, ok := m.channels[name]
	return channel, ok
}

func (m *Manager) HasChannel(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.channels[name]
	return ok
}

func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]interface{})
	for name, channel := range m.channels {
		status[name] = map[string]interface{}{
			"enabled": true,
			"running": channel.IsRunning(),
		}
	}
	return status
}

func (m *Manager) GetEnabledChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.channels))
	for name := range m.channels {
		names = append(names, name)
	}
	return names
}

func (m *Manager) RegisterChannel(name string, channel base.Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[name] = channel
}

func (m *Manager) UnregisterChannel(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.channels, name)
}

func (m *Manager) SendToChannel(ctx context.Context, channelName, chatID, content string) error {
	m.mu.RLock()
	channel, exists := m.channels[channelName]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("channel %s not found", channelName)
	}

	msg := bus.OutboundMessage{
		Channel: channelName,
		ChatID:  chatID,
		Content: content,
	}

	return channel.Send(ctx, msg)
}
