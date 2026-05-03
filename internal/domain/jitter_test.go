package domain

import (
	"sync"
	"testing"
)

func TestBuffer_PopBeforeMin(t *testing.T) {
	t.Parallel()
	b := NewBuffer(3)
	b.Push(&Packet{Sequence: 1})
	b.Push(&Packet{Sequence: 2})
	if got := b.Pop(); got != nil {
		t.Fatalf("pop before minCount = %+v, want nil", got)
	}
}

func TestBuffer_PushPopOrdered(t *testing.T) {
	t.Parallel()
	b := NewBuffer(2)
	b.Push(&Packet{Sequence: 1})
	b.Push(&Packet{Sequence: 2})
	b.Push(&Packet{Sequence: 3})
	if got := b.Pop(); got == nil || got.Sequence != 1 {
		t.Fatalf("pop[0] = %+v, want seq=1", got)
	}
	if got := b.Pop(); got == nil || got.Sequence != 2 {
		t.Fatalf("pop[1] = %+v, want seq=2", got)
	}
	if got := b.Pop(); got != nil {
		t.Fatalf("pop[2] = %+v, want nil (below minCount)", got)
	}
}

func TestBuffer_PushReorders(t *testing.T) {
	t.Parallel()
	b := NewBuffer(1)
	b.Push(&Packet{Sequence: 5})
	b.Push(&Packet{Sequence: 2})
	b.Push(&Packet{Sequence: 9})
	b.Push(&Packet{Sequence: 1})

	var got []uint32
	for {
		p := b.Pop()
		if p == nil {
			break
		}
		got = append(got, p.Sequence)
	}
	want := []uint32{1, 2, 5, 9}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("pop order = %v, want %v", got, want)
		}
	}
}

func TestBuffer_Len(t *testing.T) {
	t.Parallel()
	b := NewBuffer(2)
	if b.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", b.Len())
	}
	b.Push(&Packet{Sequence: 1})
	if b.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", b.Len())
	}
	b.Push(&Packet{Sequence: 2})
	if b.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", b.Len())
	}
	_ = b.Pop()
	if b.Len() != 1 {
		t.Fatalf("Len() after pop = %d, want 1", b.Len())
	}
}

// TestBuffer_ConcurrentPushPop exercises the Buffer mutex under -race;
// AGENTS.md mandates race testing of jitter paths.
func TestBuffer_ConcurrentPushPop(t *testing.T) {
	t.Parallel()
	b := NewBuffer(4)
	var wg sync.WaitGroup

	const writers uint32 = 4
	const perWriter uint32 = 100
	wg.Add(int(writers))
	for w := uint32(0); w < writers; w++ {
		base := w * perWriter
		go func() {
			defer wg.Done()
			for i := uint32(0); i < perWriter; i++ {
				b.Push(&Packet{Sequence: base + i})
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var popped uint32
		for popped < writers*perWriter-4 {
			if p := b.Pop(); p != nil {
				popped++
			}
		}
	}()

	wg.Wait()
}
