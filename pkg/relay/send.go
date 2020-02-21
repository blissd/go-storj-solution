package relay

import (
	"fmt"
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
	bs, err := EncodeString(msgFileName, name)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}

func (s *Session) RecvFileName() (string, error) {
	hdr := make([]byte, 2)
	_, err := s.conn.Read(hdr)
	if err != nil {
		return "", err
	}
	if hdr[0] != msgFileName {
		return "", fmt.Errorf("expected %v, got %v", msgFileName, hdr[0])
	}
	length := hdr[1]
	name := make([]byte, length)
	_, err = s.conn.Read(name)
	if err != nil {
		return "", err
	}

	return string(name), nil
}

func (s *Session) SendSecret(secret string) error {
	bs, err := EncodeString(msgSecretCode, secret)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}

func (s *Session) SendFileLength(length uint32) error {
	bs, err := EncodeUint32(msgFileLength, length)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}
