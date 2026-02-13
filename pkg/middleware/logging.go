package middleware

import (
"log"
"time"
)

// Logger handles logging for audio streams.
// It aggregates logs to prevent flooding stdout (200 packets/sec).
type Logger struct {
	lastLog    time.Time
	byteCount  int
	ktPackets  int
	direction  string
}

func New(direction string) *Logger {
	return &Logger{
		direction: direction,
		lastLog:   time.Now(),
	}
}

func (l *Logger) Log(bytes int, addr string) {
	l.byteCount += bytes
	l.ktPackets++
	
	if time.Since(l.lastLog) >= 5*time.Second {
		log.Printf("[%s] Addr: %s | Rate: %d pkts/5s | Bytes: %d", l.direction, addr, l.ktPackets, l.byteCount)
		l.lastLog = time.Now()
		l.byteCount = 0
		l.ktPackets = 0
	}
}
