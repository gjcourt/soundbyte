package stdin

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestSource_ReadFrame_FromBytes(t *testing.T) {
	t.Parallel()
	want := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	s := &Source{r: bufio.NewReader(bytes.NewReader(want))}
	got := make([]byte, len(want))
	if err := s.ReadFrame(got); err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ReadFrame got = %x, want %x", got, want)
	}
}

func TestSource_ReadFrame_EOF(t *testing.T) {
	t.Parallel()
	s := &Source{r: bufio.NewReader(bytes.NewReader(nil))}
	got := make([]byte, 4)
	err := s.ReadFrame(got)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("ReadFrame on empty = %v, want io.EOF", err)
	}
}

func TestSource_ReadFrame_Partial(t *testing.T) {
	t.Parallel()
	// 3 bytes available, asking for 6 — io.ReadFull yields ErrUnexpectedEOF.
	s := &Source{r: bufio.NewReader(bytes.NewReader([]byte{1, 2, 3}))}
	got := make([]byte, 6)
	err := s.ReadFrame(got)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("partial ReadFrame = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestNewFileSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "pcm.bin")
	want := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}
	s, err := NewFileSource(path)
	if err != nil {
		t.Fatalf("NewFileSource: %v", err)
	}
	got := make([]byte, len(want))
	if err := s.ReadFrame(got); err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ReadFrame got = %x, want %x", got, want)
	}
}

func TestNewFileSource_Missing(t *testing.T) {
	t.Parallel()
	_, err := NewFileSource(filepath.Join(t.TempDir(), "no-such-file"))
	if err == nil {
		t.Fatalf("NewFileSource on missing path returned nil error")
	}
}
