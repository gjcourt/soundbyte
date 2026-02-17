// Package auth provides optional HMAC-SHA256 packet authentication for
// soundbyte. When a shared token is configured, the sender appends a 32-byte
// HMAC to each packet and the receiver verifies it before processing.
//
// Wire format (when auth is enabled):
//
//	[original packet bytes][32-byte HMAC-SHA256]
//
// When auth is disabled the packet is sent/received as-is.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

// MACSize is the length in bytes of the HMAC-SHA256 tag.
const MACSize = 32

// ErrInvalidMAC is returned when the HMAC on a received packet doesn't match.
var ErrInvalidMAC = errors.New("invalid HMAC")

// Sign appends an HMAC-SHA256 tag to data using the given key.
// If key is nil, data is returned unchanged (auth disabled).
func Sign(data, key []byte) []byte {
	if key == nil {
		return data
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return append(data, mac.Sum(nil)...)
}

// Verify checks and strips the HMAC-SHA256 tag from data.
// If key is nil, data is returned unchanged (auth disabled).
// Returns the original payload (without the tag) on success.
func Verify(data, key []byte) ([]byte, error) {
	if key == nil {
		return data, nil
	}
	if len(data) < MACSize {
		return nil, ErrInvalidMAC
	}
	payload := data[:len(data)-MACSize]
	tag := data[len(data)-MACSize:]

	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	expected := mac.Sum(nil)

	if !hmac.Equal(tag, expected) {
		return nil, ErrInvalidMAC
	}
	return payload, nil
}
