// Package app holds the application services for soundbyte.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"soundbyte/internal/domain"
	"soundbyte/internal/ports/inbound"
	"soundbyte/internal/ports/outbound"
)

// streamingService implements inbound.StreamingService.
type streamingService struct {
	source outbound.PCMSource
	sender outbound.PacketSender
}

// NewStreamingService creates a StreamingService that reads PCM from source
// and sends encoded packets via sender.
func NewStreamingService(source outbound.PCMSource, sender outbound.PacketSender) inbound.StreamingService {
	return &streamingService{source: source, sender: sender}
}

// Stream reads PCM frames and sends encoded packets until EOF or ctx is done.
func (s *streamingService) Stream(ctx context.Context) error {
	pcm := make([]byte, domain.FrameSizeBytes)
	seq := uint32(0)

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := s.source.ReadFrame(pcm); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("reading PCM: %w", err)
		}

		payload := make([]byte, len(pcm))
		copy(payload, pcm)

		pkt := &domain.Packet{
			Sequence:  seq,
			Timestamp: uint64(time.Now().UnixNano()),
			Data:      payload,
		}
		seq++

		encoded, err := pkt.Encode()
		if err != nil {
			continue
		}

		if _, err := s.sender.Send(encoded); err != nil {
			continue
		}
	}
}
