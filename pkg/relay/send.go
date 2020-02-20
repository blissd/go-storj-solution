package relay

import "net"

type state byte

const (
	initial state = 1
	started state = 2
	sending state = 3
)

type SendSession struct {
	conn net.Conn
}

func (s *SendSession) Close() error {
	return s.conn.Close()
}

func New(addr string) (*SendSession, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &SendSession{conn: conn}, nil
}
