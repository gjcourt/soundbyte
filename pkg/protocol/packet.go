// Package protocol defines the UDP packet format for the audio streamer.
package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	// HeaderSize is the size of the packet header in bytes:
	// 4 bytes Sequence + 8 bytes Timestamp.
	HeaderSize = 12
	// MaxPacketSize is the maximum size of a packet.
	// We try to keep under traditional MTU (1500).
	MaxPacketSize = 1400
)

var (
	// ErrPacketTooShort is returned when decoding a packet that is smaller than the HeaderSize.
	ErrPacketTooShort = errors.New("packet too short")
)

// Packet represents a single unit of audio data sent over the network.
type Packet struct {
	// Sequence is a monotonic counter for ordering.
	Sequence uint32
	// Timestamp is the creation time in nanoseconds.
	Timestamp uint64
	// Data is the raw audio payload (PCM).
	Data []byte
}

// Encode serializes the Packet into a byte slice.
func (p *Packet) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	// Pre-allocate minimal size
	buf.Grow(HeaderSize + len(p.Data))

	if err := binary.Write(buf, binary.BigEndian, p.Sequence); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, p.Timestamp); err != nil {
		return nil, err
	}
	if _, err := buf.Write(p.Data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode deserializes a byte slice into a Packet.
func Decode(data []byte) (*Packet, error) {
	if len(data) < HeaderSize {
		return nil, ErrPacketTooShort
	}

	buf := bytes.NewReader(data)
	var seq uint32
	if err := binary.Read(buf, binary.BigEndian, &seq); err != nil {
		return nil, err
	}
	var ts uint64
	if err := binary.Read(buf, binary.BigEndian, &ts); err != nil {
		return nil, err
	}

	// Read remaining logic
	payload := make([]byte, buf.Len())
	if _, err := io.ReadFull(buf, payload); err != nil {
		return nil, err
	}

	return &Packet{
		Sequence:  seq,
		Timestamp: ts,
		Data:      payload,
	}, nil
}
