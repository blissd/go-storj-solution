package client

import (
	"fmt"
	"go-storj-solution/pkg/wire"
	"io"
	"net"
)

const (
	// MsgSend identifiers sender
	MsgSend Side = iota + 1

	// MsgRecv identifies receiver
	MsgRecv
)

// Side of a transfer
type Side byte

type SendRequest struct {
	// Body of file to send
	Body io.Reader

	// Name of file to send
	Name string

	// Length of file to send
	Length int64
}

type SendResponse struct {
	Secret string
	Errors <-chan error
}

type RecvResponse struct {
	Body io.ReadCloser
	Name string
}

//Service for clients to send and receive files through the relay proxy
type Service interface {
	// Send sends files through the relay proxy.
	// A secret and control channel is returned. Receives will need to provide the secret
	// to receive the file. The control channel is used to report errors and notify the
	// caller that the file has been sent. If the channel closes and doesn't contain an error then
	// the file was successfully sent. If an error occurred then it will be available on the channel.
	Send(request *SendRequest) (*SendResponse, error)

	// Recv receives files through the relay proxy. Files can only be received with
	// the correct secret. If the secret is valid, then a reader to stream the file is returned
	// and also a file name.
	Recv(secret string) (*RecvResponse, error)

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

func (s *service) Send(r *SendRequest) (*SendResponse, error) {
	// Tell relay proxy we are the sender
	if err := s.enc.EncodeByte(byte(MsgSend)); err != nil {
		return nil, fmt.Errorf("sending msg send byte: %w", err)
	}

	// Receive secret from relay proxy
	secret, err := s.dec.DecodeString()
	if err != nil {
		return nil, fmt.Errorf("receiving secret: %w", err)
	}

	errs := make(chan error, 1)

	response := &SendResponse{
		Secret: secret,
		Errors: errs,
	}

	go func() {
		defer close(errs)

		// Wait for receiver to join relay proxy
		if b, err := s.dec.DecodeByte(); b != byte(MsgRecv) || err != nil {
			errs <- fmt.Errorf("bad receiver [%v]: %w", b, err)
			return
		}

		// Send file name
		if err := s.enc.EncodeString(r.Name); err != nil {
			errs <- fmt.Errorf("sending file name: %w", err)
			return
		}

		if err := s.enc.EncodeInt64(r.Length); err != nil {
			errs <- fmt.Errorf("sending length: %w", err)
			return
		}

		written, err := io.Copy(s.con, r.Body)
		if err != nil {
			errs <- fmt.Errorf("sending file: %w", err)
			return
		}
		if written != r.Length {
			errs <- fmt.Errorf("unexpected length [%v]", written)
			return
		}
	}()

	return response, nil
}

func (s *service) Recv(secret string) (*RecvResponse, error) {
	if err := s.enc.EncodeByte(byte(MsgRecv)); err != nil {
		return nil, fmt.Errorf("sending msg recv byte: %w", err)
	}

	// send secret
	if err := s.enc.EncodeString(secret); err != nil {
		return nil, fmt.Errorf("sending secret: %w", err)
	}

	// receive file name
	name, err := s.dec.DecodeString()
	if err != nil {
		return nil, fmt.Errorf("receiving file name: %w", err)
	}

	// Receive file length
	length, err := s.dec.DecodeInt64()
	if err != nil {
		return nil, fmt.Errorf("receiving file length: %w", err)
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

	response := &RecvResponse{
		Body: pr,
		Name: name,
	}

	return response, nil
}

func (s *service) Close() error {
	return s.con.Close()
}
