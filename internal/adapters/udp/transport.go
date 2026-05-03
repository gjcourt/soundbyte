// Package udp provides UDP-based packet sender and receiver adapters.
package udp

import (
	"net"

	"soundbyte/internal/ports/outbound"
	"soundbyte/pkg/auth"
)

// Sender sends encoded (and optionally signed) packets over UDP.
type Sender struct {
	conn    *net.UDPConn
	authKey []byte
}

var _ outbound.PacketSender = (*Sender)(nil)

// NewSender creates a UDP Sender. Pass authKey=nil to disable packet signing.
func NewSender(conn *net.UDPConn, authKey []byte) *Sender {
	return &Sender{conn: conn, authKey: authKey}
}

// Send signs (if key is set) and writes data to the UDP connection.
func (s *Sender) Send(data []byte) (int, error) {
	data = auth.Sign(data, s.authKey)
	return s.conn.Write(data)
}

// Receiver receives raw bytes from a UDP connection, verifying the HMAC if a
// key is configured.
type Receiver struct {
	conn    *net.UDPConn
	authKey []byte
}

var _ outbound.PacketReceiver = (*Receiver)(nil)

// NewReceiver creates a UDP Receiver. Pass authKey=nil to disable verification.
func NewReceiver(conn *net.UDPConn, authKey []byte) *Receiver {
	return &Receiver{conn: conn, authKey: authKey}
}

// Receive reads one UDP datagram, verifies auth, and returns the payload and
// sender address string.
func (r *Receiver) Receive() ([]byte, string, error) {
	buf := make([]byte, 2048)
	n, raddr, err := r.conn.ReadFromUDP(buf)
	if err != nil {
		return nil, "", err
	}
	data := make([]byte, n)
	copy(data, buf[:n])

	verified, err := auth.Verify(data, r.authKey)
	if err != nil {
		return nil, raddr.String(), err
	}
	return verified, raddr.String(), nil
}
