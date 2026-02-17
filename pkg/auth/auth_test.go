package auth

import (
	"bytes"
	"testing"
)

func TestSign_NilKey(t *testing.T) {
	data := []byte("hello")
	got := Sign(data, nil)
	if !bytes.Equal(got, data) {
		t.Fatalf("expected unchanged data, got %x", got)
	}
}

func TestVerify_NilKey(t *testing.T) {
	data := []byte("hello")
	got, err := Verify(data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("expected unchanged data, got %x", got)
	}
}

func TestSignAndVerify(t *testing.T) {
	key := []byte("my-secret-token")
	data := []byte("packet-payload-here")

	signed := Sign(data, key)
	if len(signed) != len(data)+MACSize {
		t.Fatalf("signed length = %d, want %d", len(signed), len(data)+MACSize)
	}

	payload, err := Verify(signed, key)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !bytes.Equal(payload, data) {
		t.Fatalf("payload = %x, want %x", payload, data)
	}
}

func TestVerify_WrongKey(t *testing.T) {
	key1 := []byte("correct")
	key2 := []byte("wrong")

	signed := Sign([]byte("data"), key1)
	_, err := Verify(signed, key2)
	if err != ErrInvalidMAC {
		t.Fatalf("expected ErrInvalidMAC, got %v", err)
	}
}

func TestVerify_TooShort(t *testing.T) {
	key := []byte("key")
	_, err := Verify([]byte("short"), key)
	if err != ErrInvalidMAC {
		t.Fatalf("expected ErrInvalidMAC, got %v", err)
	}
}

func TestVerify_Tampered(t *testing.T) {
	key := []byte("key")
	signed := Sign([]byte("original"), key)

	// Tamper with the payload
	signed[0] ^= 0xFF

	_, err := Verify(signed, key)
	if err != ErrInvalidMAC {
		t.Fatalf("expected ErrInvalidMAC, got %v", err)
	}
}
