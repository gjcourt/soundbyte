// Package main implements the audio client which receives UDP audio and plays it.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"soundbyte/pkg/jitter"
	"soundbyte/pkg/middleware"
	"soundbyte/pkg/protocol"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const (
	SampleRate = 48000
	Channels   = 2
)

// PCMStreamer implements beep.Streamer for Raw PCM
type PCMStreamer struct {
	jb *jitter.Buffer

	// Buffer for one frame (bytes)
	rawBytes []byte
	// Current read position in rawBytes
	rawPos int
}

func NewPCMStreamer(jb *jitter.Buffer) *PCMStreamer {
	return &PCMStreamer{
		jb:       jb,
		rawBytes: nil,
		rawPos:   0,
	}
}

// Stream fills samples with audio.
func (s *PCMStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	filled := 0
	for filled < len(samples) {
		// If we exhausted current PCM buffer, fetch next packet
		if s.rawBytes == nil || s.rawPos >= len(s.rawBytes) {
			pkt := s.jb.Pop()
			if pkt == nil {
				// No data available.
				break
			}
			s.rawBytes = pkt.Data
			s.rawPos = 0
		}

		// Copy from rawBytes to samples
		// rawBytes is []byte (S16LE interleaved)
		// Need 4 bytes per stereo sample (2 bytes L + 2 bytes R)
		for s.rawPos+4 <= len(s.rawBytes) && filled < len(samples) {

			// Little Endian Int16 -> Float64
			lInt := int16(binary.LittleEndian.Uint16(s.rawBytes[s.rawPos : s.rawPos+2]))
			rInt := int16(binary.LittleEndian.Uint16(s.rawBytes[s.rawPos+2 : s.rawPos+4]))

			samples[filled][0] = float64(lInt) / 32768.0
			samples[filled][1] = float64(rInt) / 32768.0

			s.rawPos += 4
			filled++
		}
	}

	return filled, true
}

func (s *PCMStreamer) Err() error {
	return nil
}

func main() {
	port := flag.Int("port", 5004, "UDP port to listen on")
	bufferPackets := flag.Int("buf", 20, "Jitter buffer size (packets). 20 * 5ms = 100ms")
	flag.Parse()

	// 1. Setup UDP
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", *port))
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	if err := conn.SetReadBuffer(1024 * 1024); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s...", addr)

	// 2. Setup Jitter Buffer
	jb := jitter.New(*bufferPackets)

	// 3. Receive Loop (Background)
	go func() {
		mw := middleware.New("RX")
		// Max UDP payload expected ~960 + header (~1000)
		buf := make([]byte, 2048)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Read error: %v", err)
				continue
			}

			mw.Log(n, addr.String())

			// Copy data
			data := make([]byte, n)
			copy(data, buf[:n])

			pkt, err := protocol.Decode(data)
			if err != nil {
				log.Printf("Packet decode error: %v", err)
				continue
			}
			jb.Push(pkt)
		}
	}()

	// 4. Setup Audio
	sr := beep.SampleRate(SampleRate)
	err = speaker.Init(sr, sr.N(time.Second/10)) // 100ms internal buffer
	if err != nil {
		log.Fatal(err)
	}

	streamer := NewPCMStreamer(jb)
	speaker.Play(streamer)

	log.Println("Client started. Playing Raw PCM...")
	select {}
}
