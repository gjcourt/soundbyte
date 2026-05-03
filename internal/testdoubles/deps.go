package testdoubles

import (
	"soundbyte/internal/ports/outbound"
)

// FakePCMSource is a function-field fake for outbound.PCMSource.
type FakePCMSource struct {
	ReadFrameFn func(buf []byte) error
}

var _ outbound.PCMSource = (*FakePCMSource)(nil)

func (f *FakePCMSource) ReadFrame(buf []byte) error {
	if f.ReadFrameFn != nil {
		return f.ReadFrameFn(buf)
	}
	return nil
}

// FakePacketSender is a function-field fake for outbound.PacketSender.
type FakePacketSender struct {
	SendFn func(data []byte) (int, error)
}

var _ outbound.PacketSender = (*FakePacketSender)(nil)

func (f *FakePacketSender) Send(data []byte) (int, error) {
	if f.SendFn != nil {
		return f.SendFn(data)
	}
	return len(data), nil
}

// FakePacketReceiver is a function-field fake for outbound.PacketReceiver.
type FakePacketReceiver struct {
	ReceiveFn func() ([]byte, string, error)
}

var _ outbound.PacketReceiver = (*FakePacketReceiver)(nil)

func (f *FakePacketReceiver) Receive() ([]byte, string, error) {
	if f.ReceiveFn != nil {
		return f.ReceiveFn()
	}
	return nil, "", nil
}

// ServerDeps aggregates all outbound-port fakes for unit tests.
type ServerDeps struct {
	Source   *FakePCMSource
	Sender   *FakePacketSender
	Receiver *FakePacketReceiver
}

// NewServerDeps returns a ServerDeps with all fakes initialised to safe zero-value defaults.
func NewServerDeps() *ServerDeps {
	return &ServerDeps{
		Source:   &FakePCMSource{},
		Sender:   &FakePacketSender{},
		Receiver: &FakePacketReceiver{},
	}
}
