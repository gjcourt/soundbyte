// Package main implements the audio client which receives UDP audio and plays it.
package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	adaptudp "soundbyte/internal/adapters/udp"
	"soundbyte/internal/domain"
	"soundbyte/pkg/middleware"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// PCMStreamer implements beep.Streamer for raw PCM pulled from the jitter buffer.
type PCMStreamer struct {
	jb *domain.Buffer

	rawBytes []byte
	rawPos   int
}

// NewPCMStreamer returns a PCMStreamer backed by jb.
func NewPCMStreamer(jb *domain.Buffer) *PCMStreamer {
	return &PCMStreamer{jb: jb}
}

// Stream fills samples with S16LE stereo audio from the jitter buffer.
func (s *PCMStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	filled := 0
	for filled < len(samples) {
		if s.rawBytes == nil || s.rawPos >= len(s.rawBytes) {
			pkt := s.jb.Pop()
			if pkt == nil {
				break
			}
			s.rawBytes = pkt.Data
			s.rawPos = 0
		}
		for s.rawPos+4 <= len(s.rawBytes) && filled < len(samples) {
			left := int16(binary.LittleEndian.Uint16(s.rawBytes[s.rawPos : s.rawPos+2]))
			right := int16(binary.LittleEndian.Uint16(s.rawBytes[s.rawPos+2 : s.rawPos+4]))
			samples[filled][0] = float64(left) / 32768.0
			samples[filled][1] = float64(right) / 32768.0
			s.rawPos += 4
			filled++
		}
	}
	return filled, true
}

// Err returns the streamer's terminal error, if any.
func (s *PCMStreamer) Err() error { return nil }

func main() {
	port := flag.Int("port", 5004, "UDP port to listen on")
	bufferPackets := flag.Int("buf", 20, "Jitter buffer size (packets). 20 * 5ms = 100ms")
	token := flag.String("token", "", "Shared secret for HMAC-SHA256 packet authentication (optional)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	var authKey []byte
	if *token != "" {
		authKey = []byte(*token)
		logger.Info("packet authentication enabled")
	}

	// Cancel the receive loop on SIGINT/SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Setup UDP
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logger.Error("resolve udp addr", "err", err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error("listen udp", "err", err)
		os.Exit(1)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetReadBuffer(1024 * 1024); err != nil {
		logger.Error("set read buffer", "err", err)
		os.Exit(1)
	}

	logger.Info("listening", "addr", addr.String())

	// 2. Setup Jitter Buffer
	jb := domain.NewBuffer(*bufferPackets)
	receiver := adaptudp.NewReceiver(conn, authKey)

	// 3. Receive Loop (Background) — exits when ctx is cancelled.
	var wg sync.WaitGroup
	wg.Add(1)
	go receiveLoop(ctx, &wg, logger, receiver, jb, conn)

	// 4. Setup Audio
	sr := beep.SampleRate(domain.SampleRate)
	if err := speaker.Init(sr, sr.N(time.Second/10)); err != nil {
		logger.Error("speaker init", "err", err)
		cancel()
		_ = conn.Close()
		wg.Wait()
		os.Exit(1)
	}

	streamer := NewPCMStreamer(jb)
	speaker.Play(streamer)

	logger.Info("client started, playing raw PCM")

	// Block until ctx is cancelled (signal received), then drain receive loop.
	<-ctx.Done()
	logger.Info("shutting down")
	// Closing the conn unblocks any in-flight ReadFromUDP.
	_ = conn.Close()
	wg.Wait()
}

// receiveLoop reads packets, decodes them, and pushes onto the jitter buffer
// until ctx is cancelled or the connection is closed.
func receiveLoop(
	ctx context.Context,
	wg *sync.WaitGroup,
	logger *slog.Logger,
	receiver *adaptudp.Receiver,
	jb *domain.Buffer,
	conn *net.UDPConn,
) {
	defer wg.Done()
	mw := middleware.New("RX")
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		// Honour ctx between blocking reads.
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		data, raddr, err := receiver.Receive()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			var nerr net.Error
			if errors.As(err, &nerr) && nerr.Timeout() {
				continue
			}
			// Drop unauthenticated or malformed packets silently.
			continue
		}
		mw.Log(len(data), raddr)

		pkt, err := domain.Decode(data)
		if err != nil {
			logger.Warn("packet decode error", "err", err)
			continue
		}
		jb.Push(pkt)
	}
}
