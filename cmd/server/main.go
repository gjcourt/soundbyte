// Package main implements the audio server which streams PCM data via UDP.
package main

import (
	"context"
	"flag"
	"log"
	"net"

	"soundbyte/internal/adapters/stdin"
	adaptudp "soundbyte/internal/adapters/udp"
	"soundbyte/internal/app"
	"soundbyte/internal/ports/outbound"
	"soundbyte/pkg/middleware"
)

func main() {
	targetAddr := flag.String("addr", "255.255.255.255:5004", "Target UDP address")
	inputPath := flag.String("input", "stdin", "Path to input pipe/file (or 'stdin')")
	token := flag.String("token", "", "Shared secret for HMAC-SHA256 packet authentication (optional)")
	flag.Parse()

	var authKey []byte
	if *token != "" {
		authKey = []byte(*token)
		log.Println("Packet authentication enabled")
	}

	// Setup input source
	var (
		source *stdin.Source
		err    error
	)
	if *inputPath == "stdin" {
		source = stdin.NewSource()
	} else {
		source, err = stdin.NewFileSource(*inputPath)
		if err != nil {
			log.Fatalf("Failed to open input: %v", err)
		}
	}

	// Setup UDP sender
	raddr, err := net.ResolveUDPAddr("udp", *targetAddr)
	if err != nil {
		log.Fatalf("Invalid address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalf("Failed to dial UDP: %v", err)
	}
	defer func() { _ = conn.Close() }()

	mw := middleware.New("TX")
	sender := adaptudp.NewSender(conn, authKey)
	loggingSender := &loggingPacketSender{inner: sender, mw: mw, addr: *targetAddr}

	svc := app.NewStreamingService(source, loggingSender)

	log.Printf("Streaming Raw PCM to %s (Expected: S16LE Stereo 48kHz)", *targetAddr)
	if err := svc.Stream(context.Background()); err != nil {
		log.Fatalf("Streaming stopped: %v", err)
	}
	log.Println("End of stream")
}

// loggingPacketSender wraps a PacketSender and logs each sent packet.
type loggingPacketSender struct {
	inner outbound.PacketSender
	mw    *middleware.Logger
	addr  string
}

func (l *loggingPacketSender) Send(data []byte) (int, error) {
	n, err := l.inner.Send(data)
	if err == nil {
		l.mw.Log(n, l.addr)
	}
	return n, err
}
