package domain

import (
	"sort"
	"sync"
)

// Buffer is a thread-safe packet reordering buffer that holds packets until
// minCount is reached, then releases them in sequence order.
type Buffer struct {
	mu       sync.Mutex
	packets  []*Packet
	minCount int
}

// NewBuffer creates a new jitter Buffer.
// minCount is the number of packets to accumulate before popping.
func NewBuffer(minCount int) *Buffer {
	return &Buffer{
		minCount: minCount,
		packets:  make([]*Packet, 0, minCount*2),
	}
}

// Push adds a packet to the buffer and keeps packets sorted by sequence number.
func (b *Buffer) Push(p *Packet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.packets = append(b.packets, p)
	sort.Slice(b.packets, func(i, j int) bool {
		return b.packets[i].Sequence < b.packets[j].Sequence
	})
}

// Pop returns the next packet if the buffer has at least minCount packets.
// Returns nil when the buffer is not full enough.
func (b *Buffer) Pop() *Packet {
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
