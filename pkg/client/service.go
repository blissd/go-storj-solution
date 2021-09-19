package client

import (
	"fmt"
	"go-storj-solution/pkg/wire"
	"io"
)

const (
	// MsgSend identifiers sender
	MsgSend Side = 'S'

	// MsgRecv identifies receiver
	MsgRecv Side = 'R'
)

// Side of a transfer
type Side byte

func (s Side) String() string {
	switch s {
	case MsgSend:
		return "sender"
	case MsgRecv:
		return "receiver"
	default:
		return fmt.Sprintf("unknown [%v]", byte(s))
	}
}

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
	Body io.Reader
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
}

//service client service
type service struct {
	enc wire.Encoder
	dec wire.Decoder
}

//NewService creates a new client service
func NewService(enc wire.Encoder, dec wire.Decoder) Service {
	return &service{
		enc: enc,
		dec: dec,
	}
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

		// Send file body
		if err := s.enc.EncodeReader(r.Body, r.Length); err != nil {
			errs <- fmt.Errorf("sending body: %w", err)
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

	r, err := s.dec.DecodeReader()
	if err != nil {
		return nil, fmt.Errorf("receiving body: %w", err)
	}

	response := &RecvResponse{
		Body: r,
		Name: name,
	}

	return response, nil
}
