package session

import (
	"fmt"
	"github.com/blissd/golang-storj-solution/pkg/wire"
	"io"
	"net"
)

const (

	// Encoded sender type
	MsgSend byte = iota + 1

	// Encoded receiver type
	MsgRecv
)

type Session struct {
	conn net.Conn
	enc  wire.FrameEncoder
	dec  wire.FrameDecoder
}

func (s *Session) Close() error {
	return s.conn.Close()
}

func New(addr string) (*Session, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}
	return &Session{
		conn: conn,
		enc:  wire.NewEncoder(conn),
		dec:  wire.NewDecoder(conn),
	}, nil
}

func Attach(conn net.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) SendFileName(name string) error {
	return s.enc.EncodeString(name)
}

func (s *Session) RecvFileName() (string, error) {
	v, err := s.recvString()
	return v, err
}

func (s *Session) SendSecret(secret string) error {
	return s.sendString(secret)
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

func (s *Session) Send(r io.Reader) (int64, error) {
	return io.Copy(s.conn, r)
}

func (s *Session) Recv(w io.Writer) (int64, error) {
	return io.Copy(w, s.conn)
}

func (s *Session) sendString(v string) error {
	return s.enc.EncodeString(v)
}

func (s *Session) recvString() (string, error) {
	return s.dec.DecodeString()
}
