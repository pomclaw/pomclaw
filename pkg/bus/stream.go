package bus

import "context"

// Streamer pushes incremental updates for a single logical response.
type Streamer interface {
	Update(ctx context.Context, content string) error
	Finalize(ctx context.Context, content string) error
	Cancel(ctx context.Context)
	HasPosted() bool
}

// StreamDelegate resolves a channel-specific streamer.
type StreamDelegate interface {
	GetStreamer(ctx context.Context, channel, chatID string) (Streamer, bool)
}
