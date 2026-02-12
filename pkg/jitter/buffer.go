// Package jitter implements a simple packet reordering buffer.
package jitter

import (
	"soundbyte/pkg/protocol"
	"sort"
	"sync"
)

// Buffer is a simple jitter buffer that reorders packets based on their sequence number.
// It is thread-safe.
type Buffer struct {
	mu       sync.Mutex
	packets  []*protocol.Packet
	minCount int
}

// New creates a new Jitter Buffer.
// minCount is the number of packets to buffer before popping.
func New(minCount int) *Buffer {
	return &Buffer{
		minCount: minCount,
		packets:  make([]*protocol.Packet, 0, minCount*2),
	}
}

// Push adds a packet to the buffer and sorts the internal queue.
func (b *Buffer) Push(p *protocol.Packet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.packets = append(b.packets, p)

	// Sort by sequence number.
	// Note: This naive sort fails on uint32 overflow. acceptable for proto.
	sort.Slice(b.packets, func(i, j int) bool {
		return b.packets[i].Sequence < b.packets[j].Sequence
	})
}

// Pop returns the next packet if we satisfy the buffering condition.
// Returns nil if the buffer is not full enough.
func (b *Buffer) Pop() *protocol.Packet {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.packets) < b.minCount {
		return nil
	}

	p := b.packets[0]
	b.packets = b.packets[1:]
	return p
}

// Len returns the current number of packets in the buffer.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.packets)
}
