// Package inbound holds driving-port interfaces for application use cases.
package inbound

import "context"

// StreamingService is the driving port for the server's PCM streaming use case.
type StreamingService interface {
	// Stream reads PCM from a source and sends encoded packets until the
	// source is exhausted or ctx is cancelled.
	Stream(ctx context.Context) error
}
