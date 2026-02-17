package middleware

import (
	"bytes"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	l := New("TEST")

	l.Log(100, "127.0.0.1")
	if l.byteCount != 100 {
		t.Errorf("Expected 100 bytes, got %d", l.byteCount)
	}
}

func TestLoggerExamples(t *testing.T) {
	// A simple run
	l := New("CLIENT")
	l.Log(1024, "localhost")
}
