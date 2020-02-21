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
	bs, err := EncodeFileName(name)
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
	bs, err := EncodeSecret(secret)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}

func (s *Session) SendFileLength(length uint32) error {
	bs, err := EncodeFileLength(length)
	if err != nil {
		return err
	}

	_, err = s.conn.Write(bs)
	return err
}
