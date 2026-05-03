// Package main implements the audio client which receives UDP audio and plays it.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"soundbyte/internal/domain"
	adaptudp "soundbyte/internal/adapters/udp"
	"soundbyte/pkg/middleware"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

// PCMStreamer implements beep.Streamer for raw PCM pulled from the jitter buffer.
type PCMStreamer struct {
	jb *domain.Buffer

	rawBytes []byte
	rawPos   int
}

func NewPCMStreamer(jb *domain.Buffer) *PCMStreamer {
	return &PCMStreamer{jb: jb}
}

// Stream fills samples with S16LE stereo audio from the jitter buffer.
func (s *PCMStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	filled := 0
	for filled < len(samples) {
		if s.rawBytes == nil || s.rawPos >= len(s.rawBytes) {
			pkt := s.jb.Pop()
			if pkt == nil {
				break
			}
			s.rawBytes = pkt.Data
			s.rawPos = 0
		}
		for s.rawPos+4 <= len(s.rawBytes) && filled < len(samples) {
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

func (s *PCMStreamer) Err() error { return nil }

func main() {
	port := flag.Int("port", 5004, "UDP port to listen on")
	bufferPackets := flag.Int("buf", 20, "Jitter buffer size (packets). 20 * 5ms = 100ms")
	token := flag.String("token", "", "Shared secret for HMAC-SHA256 packet authentication (optional)")
	flag.Parse()

	var authKey []byte
	if *token != "" {
		authKey = []byte(*token)
		log.Println("Packet authentication enabled")
	}

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
	jb := domain.NewBuffer(*bufferPackets)
	receiver := adaptudp.NewReceiver(conn, authKey)

	// 3. Receive Loop (Background)
	go func() {
		mw := middleware.New("RX")
		for {
			data, raddr, err := receiver.Receive()
			if err != nil {
				// Drop unauthenticated or malformed packets silently
				continue
			}
			mw.Log(len(data), raddr)

			pkt, err := domain.Decode(data)
			if err != nil {
				log.Printf("Packet decode error: %v", err)
				continue
			}
			jb.Push(pkt)
		}
	}()

	// 4. Setup Audio
	sr := beep.SampleRate(domain.SampleRate)
	err = speaker.Init(sr, sr.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	streamer := NewPCMStreamer(jb)
	speaker.Play(streamer)

	log.Println("Client started. Playing Raw PCM...")
	select {}
}
