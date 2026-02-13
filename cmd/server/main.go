// Package main implements the audio server which streams PCM data via UDP.
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"

	"soundbyte/pkg/middleware"
	"soundbyte/pkg/protocol"
)

func main() {
	targetAddr := flag.String("addr", "255.255.255.255:5004", "Target UDP address")
	inputPath := flag.String("input", "stdin", "Path to input pipe/file (or 'stdin')")
	flag.Parse()

	// 1. Setup Input
	var input io.Reader
	if *inputPath == "stdin" {
		input = os.Stdin
	} else {
		f, err := os.Open(*inputPath)
		if err != nil {
			log.Fatalf("Failed to open input: %v", err)
		}
		defer f.Close()
		input = f
	}

	reader := bufio.NewReader(input)

	// 2. Setup Network
	raddr, err := net.ResolveUDPAddr("udp", *targetAddr)
	if err != nil {
		log.Fatalf("Invalid address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalf("Failed to dial UDP: %v", err)
	}
	defer conn.Close()

	// Config: 48kHz, S16LE, Stereo
	// We use RAW PCM instead of Opus to maintain "Pure Go" requirement (pion/opus is decode-only).
	// To fit in MTU (1500), we must use small frames.
	// 48000 Hz * 2 chan * 2 bytes = 192,000 bytes/sec.
	// 5ms = 192000 * 0.005 = 960 bytes. Perfect fit.
	const frameSizeMs = 5
	const sampleRate = 48000
	const channels = 2
	const frameSizeBytes = 192000 * frameSizeMs / 1000 // 960 bytes

	pcmBytes := make([]byte, frameSizeBytes)
	seq := uint32(0)

	log.Printf("Streaming Raw PCM to %s (Expected: S16LE Stereo 48kHz)", *targetAddr)

	mw := middleware.New("TX")

	for {
		// Read full PCM frame
		// This blocks until enough data is available (natural pacing for live sources)
		_, err := io.ReadFull(reader, pcmBytes)
		if err != nil {
			if err == io.EOF {
				log.Println("End of stream")
				break
			}
			log.Printf("Error reading input: %v", err)
			break
		}

		// Create Packet
		pkt := &protocol.Packet{
			Sequence:  seq,
			Timestamp: uint64(time.Now().UnixNano()),
			Data:      pcmBytes, // Raw PCM
		}
		seq++

		encodedBytes, err := pkt.Encode()
		if err != nil {
			log.Printf("Packet encode error: %v", err)
			continue
		}

		// Send
		n, err := conn.Write(encodedBytes)
		if err != nil {
			log.Printf("UDP write error: %v", err)
		} else {
			mw.Log(n, *targetAddr)
		}
	}
}
