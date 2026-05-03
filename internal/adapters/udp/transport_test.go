package udp

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"soundbyte/pkg/auth"
)

// loopbackPair returns a connected sender + receiver UDP pair on 127.0.0.1.
func loopbackPair(t *testing.T) (sendConn, recvConn *net.UDPConn, cleanup func()) {
	t.Helper()
	rAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("resolve recv: %v", err)
	}
	recvConn, err = net.ListenUDP("udp", rAddr)
	if err != nil {
		t.Fatalf("listen recv: %v", err)
	}
	sendConn, err = net.DialUDP("udp", nil, recvConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		_ = recvConn.Close()
		t.Fatalf("dial send: %v", err)
	}
	cleanup = func() {
		_ = sendConn.Close()
		_ = recvConn.Close()
	}
	return sendConn, recvConn, cleanup
}

func TestSendReceive_NoAuth(t *testing.T) {
	t.Parallel()
	sendConn, recvConn, cleanup := loopbackPair(t)
	defer cleanup()

	s := NewSender(sendConn, nil)
	r := NewReceiver(recvConn, nil)

	want := []byte("hello-pcm")
	if _, err := s.Send(want); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if err := recvConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	got, raddr, err := r.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Receive got = %x, want %x", got, want)
	}
	if raddr == "" {
		t.Fatalf("Receive raddr empty")
	}
}

func TestSendReceive_WithAuth(t *testing.T) {
	t.Parallel()
	sendConn, recvConn, cleanup := loopbackPair(t)
	defer cleanup()

	key := []byte("shared-secret")
	s := NewSender(sendConn, key)
	r := NewReceiver(recvConn, key)

	want := []byte("authenticated-payload")
	if _, err := s.Send(want); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if err := recvConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	got, _, err := r.Receive()
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Receive got = %x, want %x", got, want)
	}
}

func TestReceive_AuthMismatch(t *testing.T) {
	t.Parallel()
	sendConn, recvConn, cleanup := loopbackPair(t)
	defer cleanup()

	s := NewSender(sendConn, []byte("send-key"))
	r := NewReceiver(recvConn, []byte("recv-key"))

	if _, err := s.Send([]byte("payload")); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if err := recvConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	_, _, err := r.Receive()
	if !errors.Is(err, auth.ErrInvalidMAC) {
		t.Fatalf("Receive err = %v, want ErrInvalidMAC", err)
	}
}

func TestReceive_TooShortForMAC(t *testing.T) {
	t.Parallel()
	sendConn, recvConn, cleanup := loopbackPair(t)
	defer cleanup()

	// Sender does NOT auth, but receiver expects a MAC — so the raw 4-byte
	// payload arrives shorter than MACSize and Verify returns ErrInvalidMAC.
	s := NewSender(sendConn, nil)
	r := NewReceiver(recvConn, []byte("key"))

	if _, err := s.Send([]byte{1, 2, 3, 4}); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if err := recvConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	_, _, err := r.Receive()
	if !errors.Is(err, auth.ErrInvalidMAC) {
		t.Fatalf("Receive err = %v, want ErrInvalidMAC", err)
	}
}

func TestReceive_ConnClosed(t *testing.T) {
	t.Parallel()
	_, recvConn, cleanup := loopbackPair(t)
	defer cleanup()

	r := NewReceiver(recvConn, nil)
	_ = recvConn.Close()

	_, _, err := r.Receive()
	if err == nil {
		t.Fatalf("Receive on closed conn returned nil error")
	}
}
