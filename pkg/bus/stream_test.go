package bus

import (
	"context"
	"testing"
)

type testStreamer struct{}

func (s *testStreamer) Update(context.Context, string) error   { return nil }
func (s *testStreamer) Finalize(context.Context, string) error { return nil }
func (s *testStreamer) Cancel(context.Context)                 {}
func (s *testStreamer) HasPosted() bool                        { return false }

type testStreamDelegate struct {
	streamer Streamer
	ok       bool
}

func (d *testStreamDelegate) GetStreamer(_ context.Context, channel, chatID string) (Streamer, bool) {
	if !d.ok || channel == "" || chatID == "" {
		return nil, false
	}
	return d.streamer, true
}

func TestMessageBus_GetStreamer(t *testing.T) {
	mb := NewMessageBus()
	if _, ok := mb.GetStreamer(context.Background(), "mattermost", "c1"); ok {
		t.Fatal("expected no streamer before delegate is set")
	}

	want := &testStreamer{}
	mb.SetStreamDelegate(&testStreamDelegate{streamer: want, ok: true})
	got, ok := mb.GetStreamer(context.Background(), "mattermost", "c1")
	if !ok {
		t.Fatal("expected streamer to be resolved")
	}
	if got != want {
		t.Fatal("resolved streamer does not match expected instance")
	}
}
