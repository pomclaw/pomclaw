package bus

import (
	"context"
	"sync"
)

type MessageBus struct {
	inbound   chan InboundMessage
	outbound  chan OutboundMessage
	handlers  map[string]MessageHandler
	streamer  StreamDelegate
	done      chan struct{}
	closeOnce sync.Once
	mu        sync.RWMutex
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		inbound:  make(chan InboundMessage, 100),
		outbound: make(chan OutboundMessage, 100),
		handlers: make(map[string]MessageHandler),
		done:     make(chan struct{}),
	}
}

func (mb *MessageBus) PublishInbound(msg InboundMessage) {
	select {
	case mb.inbound <- msg:
	case <-mb.done:
	}
}

func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case msg := <-mb.inbound:
		return msg, true
	case <-mb.done:
		return InboundMessage{}, false
	case <-ctx.Done():
		return InboundMessage{}, false
	}
}

func (mb *MessageBus) PublishOutbound(msg OutboundMessage) {
	select {
	case mb.outbound <- msg:
	case <-mb.done:
	}
}

func (mb *MessageBus) SubscribeOutbound(ctx context.Context) (OutboundMessage, bool) {
	select {
	case msg := <-mb.outbound:
		return msg, true
	case <-mb.done:
		return OutboundMessage{}, false
	case <-ctx.Done():
		return OutboundMessage{}, false
	}
}

func (mb *MessageBus) RegisterHandler(channel string, handler MessageHandler) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.handlers[channel] = handler
}

func (mb *MessageBus) GetHandler(channel string) (MessageHandler, bool) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	handler, ok := mb.handlers[channel]
	return handler, ok
}

func (mb *MessageBus) SetStreamDelegate(delegate StreamDelegate) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	mb.streamer = delegate
}

func (mb *MessageBus) GetStreamer(ctx context.Context, channel, chatID string) (Streamer, bool) {
	mb.mu.RLock()
	delegate := mb.streamer
	mb.mu.RUnlock()
	if delegate == nil {
		return nil, false
	}
	return delegate.GetStreamer(ctx, channel, chatID)
}

func (mb *MessageBus) Close() {
	mb.closeOnce.Do(func() { close(mb.done) })
}
