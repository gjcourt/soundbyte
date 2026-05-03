package app

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"soundbyte/internal/domain"
	"soundbyte/internal/testdoubles"
)

// TestStream_HappyPath: source returns one frame, then EOF; service should
// encode and send exactly one packet and return nil.
func TestStream_HappyPath(t *testing.T) {
	t.Parallel()
	deps := testdoubles.NewServerDeps()

	var calls int32
	deps.Source.ReadFrameFn = func(buf []byte) error {
		c := atomic.AddInt32(&calls, 1)
		if c == 1 {
			for i := range buf {
				buf[i] = byte(i % 256)
			}
			return nil
		}
		return io.EOF
	}

	var sentMu sync.Mutex
	var sent [][]byte
	deps.Sender.SendFn = func(data []byte) (int, error) {
		sentMu.Lock()
		cp := make([]byte, len(data))
		copy(cp, data)
		sent = append(sent, cp)
		sentMu.Unlock()
		return len(data), nil
	}

	svc := NewStreamingService(deps.Source, deps.Sender)
	if err := svc.Stream(context.Background()); err != nil {
		t.Fatalf("Stream returned err = %v, want nil", err)
	}

	sentMu.Lock()
	defer sentMu.Unlock()
	if len(sent) != 1 {
		t.Fatalf("sent %d packets, want 1", len(sent))
	}
	if len(sent[0]) != domain.HeaderSize+domain.FrameSizeBytes {
		t.Fatalf("packet size = %d, want %d", len(sent[0]), domain.HeaderSize+domain.FrameSizeBytes)
	}
	pkt, err := domain.Decode(sent[0])
	if err != nil {
		t.Fatalf("decode sent packet: %v", err)
	}
	if pkt.Sequence != 0 {
		t.Fatalf("first packet sequence = %d, want 0", pkt.Sequence)
	}
}

// TestStream_EOFOnFirstRead: returns nil immediately on EOF.
func TestStream_EOFOnFirstRead(t *testing.T) {
	t.Parallel()
	deps := testdoubles.NewServerDeps()
	deps.Source.ReadFrameFn = func(_ []byte) error { return io.EOF }

	svc := NewStreamingService(deps.Source, deps.Sender)
	if err := svc.Stream(context.Background()); err != nil {
		t.Fatalf("Stream() = %v, want nil on EOF", err)
	}
}

// TestStream_WrappedEOF ensures errors.Is is used (not == io.EOF).
func TestStream_WrappedEOF(t *testing.T) {
	t.Parallel()
	deps := testdoubles.NewServerDeps()
	deps.Source.ReadFrameFn = func(_ []byte) error {
		return wrapped{err: io.EOF}
	}

	svc := NewStreamingService(deps.Source, deps.Sender)
	if err := svc.Stream(context.Background()); err != nil {
		t.Fatalf("Stream() = %v, want nil on wrapped EOF", err)
	}
}

// TestStream_ContextCancel: a cancelled context should make Stream return
// ctx.Err() without further reads.
func TestStream_ContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	deps := testdoubles.NewServerDeps()
	called := false
	deps.Source.ReadFrameFn = func(_ []byte) error {
		called = true
		return nil
	}

	svc := NewStreamingService(deps.Source, deps.Sender)
	err := svc.Stream(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Stream() err = %v, want context.Canceled", err)
	}
	if called {
		t.Fatalf("ReadFrame called despite cancelled ctx")
	}
}

// TestStream_ReadError: a non-EOF read error should be returned wrapped.
func TestStream_ReadError(t *testing.T) {
	t.Parallel()
	deps := testdoubles.NewServerDeps()
	want := errors.New("disk on fire")
	deps.Source.ReadFrameFn = func(_ []byte) error { return want }

	svc := NewStreamingService(deps.Source, deps.Sender)
	err := svc.Stream(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("Stream() err = %v, want wraps %v", err, want)
	}
}

// TestStream_SendErrorContinues: a transient send error should not abort
// the loop; service should keep trying.
func TestStream_SendErrorContinues(t *testing.T) {
	t.Parallel()
	deps := testdoubles.NewServerDeps()

	var reads int32
	deps.Source.ReadFrameFn = func(_ []byte) error {
		c := atomic.AddInt32(&reads, 1)
		if c >= 3 {
			return io.EOF
		}
		return nil
	}

	var sendCalls int32
	deps.Sender.SendFn = func(_ []byte) (int, error) {
		atomic.AddInt32(&sendCalls, 1)
		return 0, errors.New("network blip")
	}

	svc := NewStreamingService(deps.Source, deps.Sender)
	if err := svc.Stream(context.Background()); err != nil {
		t.Fatalf("Stream() = %v, want nil at EOF", err)
	}
	if atomic.LoadInt32(&sendCalls) != 2 {
		t.Fatalf("send calls = %d, want 2", sendCalls)
	}
}

type wrapped struct {
	err error
}

func (w wrapped) Error() string { return "wrapped: " + w.err.Error() }
func (w wrapped) Unwrap() error { return w.err }
