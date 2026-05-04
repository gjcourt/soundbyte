// Package outbound holds driven-port interfaces for I/O and transport.
package outbound

// PCMSource reads raw PCM frames from an audio source.
type PCMSource interface {
	// ReadFrame fills buf with exactly len(buf) bytes of PCM data.
	// Returns error (including io.EOF) when the source is exhausted.
	ReadFrame(buf []byte) error
}

// PacketSender sends encoded, optionally authenticated, packets over the wire.
type PacketSender interface {
	Send(data []byte) (int, error)
}

// PacketReceiver receives raw bytes from the wire and returns them.
type PacketReceiver interface {
	Receive() ([]byte, string, error)
}
