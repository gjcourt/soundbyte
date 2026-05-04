// Package domain defines the core types and business rules for soundbyte.
// It has no dependencies outside the standard library.
package domain

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	// HeaderSize is the size of the packet header in bytes (4B seq + 8B timestamp).
	HeaderSize = 12
	// MaxPacketSize is the maximum packet size (kept under traditional MTU of 1500).
	MaxPacketSize = 1400

	// SampleRate is the audio sample rate in Hz (48kHz S16LE stereo).
	SampleRate = 48000
	// Channels is the number of audio channels (stereo).
	Channels = 2
	// FrameSizeMs is the duration of one audio frame in milliseconds.
	FrameSizeMs = 5
	// FrameSizeBytes is the number of PCM bytes in one 5ms frame: 48000 * 2ch * 2bytes * 0.005s.
	FrameSizeBytes = 960
)

// ErrPacketTooShort is returned when decoding a packet shorter than HeaderSize.
var ErrPacketTooShort = errors.New("packet too short")

// Packet is a single unit of audio data sent over the network.
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
	payload := make([]byte, buf.Len())
	if _, err := io.ReadFull(buf, payload); err != nil {
		return nil, err
	}
	return &Packet{Sequence: seq, Timestamp: ts, Data: payload}, nil
}
