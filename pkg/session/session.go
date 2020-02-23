package session

import (
	"fmt"
	"github.com/blissd/golang-storj-solution/pkg/wire"
	"net"
)

const (
	// Encoded sender type
	MsgSend byte = iota + 1

	// Encoded receiver type
	MsgRecv
)

type Session struct {
	net.Conn
	enc wire.FrameEncoder
	dec wire.FrameDecoder
}

func New(addr string) (*Session, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}
	return &Session{
		Conn: conn,
		enc:  wire.NewEncoder(conn),
		dec:  wire.NewDecoder(conn),
	}, nil
}

func (s *Session) SendFileName(name string) error {
	return s.enc.EncodeString(name)
}

func (s *Session) RecvFileName() (string, error) {
	return s.dec.DecodeString()
}

func (s *Session) SendSecret(secret string) error {
	return s.enc.EncodeString(secret)
}

func (s *Session) RecvSecret() (string, error) {
	return s.dec.DecodeString()
}

func (s *Session) SendFileLength(length int64) error {
	return s.enc.EncodeInt64(length)
}

func (s *Session) RecvFileLength() (int64, error) {
	return s.dec.DecodeInt64()
}

// Informs server that client is a receiver.
// Informs sender that receiver is connected and ready.
func (s *Session) SendSendReady() error {
	return s.enc.EncodeByte(MsgSend)
}

// Informs server that client is a receiver.
// Informs sender that receiver is connected and ready.
func (s *Session) SendRecvReady() error {
	return s.enc.EncodeByte(MsgRecv)
}

// Informs server that client is a receiver.
// Informs sender that receiver is connected and ready.
func (s *Session) WaitForRecv() error {
	b, err := s.dec.DecodeByte()
	if err != nil {
		return fmt.Errorf("session.WaitForRecv: %w", err)
	}
	if b != MsgRecv {
		return fmt.Errorf("session.WaitForRecv: wrong byte: %v", b)
	}
	return nil
}
