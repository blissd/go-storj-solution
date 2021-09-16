package client

import (
	"fmt"
	"go-storj-solution/pkg/wire"
	"io"
	"net"
	"os"
)

const (
	// MsgSend identifiers sender
	MsgSend byte = iota + 1

	// MsgRecv identifies receiver
	MsgRecv
)

//Service for clients to send and receive files through the relay proxy
type Service interface {
	// Send sends files through the relay proxy.
	// A secret and control channel is returned. Receives will need to provide the secret
	// to receive the file. The control channel is used to report errors and notify the
	// caller that the file has been sent. If the channel closes and doesn't contain an error then
	// the file was successfully sent. If an error occurred then it will be available on the channel.
	Send(*os.File) (string, <-chan error)

	// Recv receives files through the relay proxy. Files can only be received with
	// the correct secret. If the secret is valid, then a reader to stream the file is returned
	// and also a file name.
	Recv(secret string) (io.Reader, string, error)

	//Close closes network connection to relay proxy.
	Close() error
}

//service client service
type service struct {
	con net.Conn
	enc wire.FrameEncoder
	dec wire.FrameDecoder
}

//NewService creates a new client service
func NewService(addr string) (Service, error) {
	con, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("new service: %w", err)
	}
	return &service{
		con: con,
		enc: wire.NewEncoder(con),
		dec: wire.NewDecoder(con),
	}, nil
}

func (s *service) Send(file *os.File) (string, <-chan error) {
	errs := make(chan error, 1)

	// format error message, send to control channel, and close channel
	fail := func(msg string, err error) <-chan error {
		errs <- fmt.Errorf("%v: %w", msg, err)
		close(errs)
		return errs
	}

	// Tell relay proxy we are the sender
	if err := s.enc.EncodeByte(MsgSend); err != nil {
		return "", fail("sending msg send byte", err)
	}

	// Receive secret from relay proxy
	secret, err := s.dec.DecodeString()
	if err != nil {
		return "", fail("receiving secret", err)
	}

	go func() {
		// Wait for receiver to join relay proxy
		if b, err := s.dec.DecodeByte(); b != MsgRecv || err != nil {
			fail(fmt.Sprintf("bad receiver [%v]", b), err)
			return
		}

		// Send file name
		if err := s.enc.EncodeString(file.Name()); err != nil {
			fail("sending file name", err)
			return
		}

		// Send file length
		info, err := file.Stat()
		if err != nil {
			fail("getting file info", err)
			return
		}

		if err := s.enc.EncodeInt64(info.Size()); err != nil {
			fail("sending length", err)
			return
		}

		written, err := io.Copy(s.con, file)
		if err != nil {
			fail("sending file", err)
			return
		}
		if written != info.Size() {
			fail(fmt.Sprintf("unexpected length [%v]", written), nil)
			return
		}
		close(errs) // signal successful copy
	}()

	return secret, errs
}

func (s *service) Recv(secret string) (io.Reader, string, error) {
	if err := s.enc.EncodeByte(MsgRecv); err != nil {
		return nil, "", fmt.Errorf("sending msg recv byte: %w", err)
	}

	// send secret
	if err := s.enc.EncodeString(secret); err != nil {
		return nil, "", fmt.Errorf("failed sending secret: %w", err)
	}

	// receive file name
	name, err := s.dec.DecodeString()
	if err != nil {
		return nil, "", fmt.Errorf("failed receiving file name: %w", err)
	}

	// Receive file length
	length, err := s.dec.DecodeInt64()
	if err != nil {
		return nil, "", fmt.Errorf("failed receiving file length: %w", err)
	}

	pr, pw := io.Pipe()
	go func() {
		copyN, err := io.Copy(pw, s.con)
		if err != nil {
			pw.CloseWithError(err)
			return
		} else if copyN != length {
			pw.CloseWithError(fmt.Errorf("unexpected file length: %d", copyN))
		} else {
			pw.Close()
		}
	}()

	return pr, name, nil
}

func (s *service) Close() error {
	return s.con.Close()
}
