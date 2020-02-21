package session

import (
	"fmt"
	"io"
	"net"
)

type state byte

const (
	initial state = 1
	started state = 2
	sending state = 3
)

const (

	// Encoded sender type
	msgSend byte = 1

	// Encoded receiver type
	msgRecv byte = 2

	// Encoded secret code
	msgSecretCode byte = 3

	// Encoded file name.
	msgFileName byte = 4

	// Encoded file length
	msgFileLength byte = 5
)

type Session struct {
	conn net.Conn
}

func (s *Session) Close() error {
	return s.conn.Close()
}

func New(addr string) (*Session, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Session{conn: conn}, nil
}

func (s *Session) SendFileName(name string) error {
	return s.sendString(msgFileName, name)
}

func (s *Session) RecvFileName() (string, error) {
	v, err := s.recvString(msgFileName)
	return v, err
}

func (s *Session) SendSecret(secret string) error {
	return s.sendString(msgSecretCode, secret)
}

func (s *Session) RecvSecret() (string, error) {
	v, err := s.recvString(msgSecretCode)
	return v, err
}

func (s *Session) SendFileLength(length uint32) error {
	bs, err := encodeUint32(msgFileLength, length)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}

func (s *Session) RecvFileLength() (uint32, error) {
	f, err := s.nextFrame()
	if err != nil {
		return 0, err
	}

	ft, v, err := decodeUint32(f)
	if err != nil {
		return 0, err
	} else if ft != msgFileLength {
		return 0, fmt.Errorf("expected %v, got %v", msgFileLength, ft)
	}
	return v, err
}

// Informs server that client is a receiver.
// Informs sender that receiver is connected and ready.
func (s *Session) SendSendReady() error {
	bs, err := encodeByte(msgSend)
	if err != nil {
		return err
	}
	_, err = s.conn.Write(bs)
	return err
}

// Informs server that client is a receiver.
// Informs sender that receiver is connected and ready.
func (s *Session) SendRecvReady() error {
	bs, err := encodeByte(msgRecv)
	if err != nil {
		return err
	}
	_, err = s.conn.Write(bs)
	return err
}

// Informs server that client is a receiver.
// Informs sender that receiver is connected and ready.
func (s *Session) WaitForRecv() error {
	bs, err := s.nextFrame()
	if err != nil {
		return err
	}
	b, err := decodeByte(bs)
	if err != nil {
		return err
	}
	if b != msgRecv {
		return fmt.Errorf("expected %v, got %v", msgRecv, b)
	}
	return nil
}

func (s *Session) Send(r io.Reader) error {
	_, err := io.Copy(s.conn, r)
	return err
}

// reads the next from from the connection
func (s *Session) nextFrame() ([]byte, error) {
	length := make([]byte, 1)
	_, err := s.conn.Read(length)
	if err != nil {
		return nil, err
	}

	frame := make([]byte, length[0]+1)
	frame[0] = length[0]
	_, err = s.conn.Read(frame[1:])
	return frame, err
}

func (s *Session) sendString(fieldType byte, v string) error {
	bs, err := encodeString(fieldType, v)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}

func (s *Session) recvString(fieldType byte) (string, error) {
	f, err := s.nextFrame()
	if err != nil {
		return "", err
	}

	ft, v, err := decodeString(f)

	if err != nil {
		return "", err
	} else if ft != fieldType {
		return "", fmt.Errorf("expected %v, got %v", fieldType, ft)
	}

	return v, nil
}
