// Package wire proves protocol framing functions.
// The data types, byte, string, and int64 can be encoded
// into frames suitable for sending across a network connection.
// The frame format is as follows:
// byte 0 - frame length, max 255 bytes
// byte 1 - frame id, effectively the type of frame.
// byte 2+ - frame payload, up to 253 bytes
// Frames don't encode any type information It is up to the caller to
// understand what type correlates to what frame.
package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// sizeOfInt64 size of int64 in bytes
const sizeOfInt64 = 8

// Encoder encodes data types to an underlying io.Writer
type Encoder interface {
	EncodeBytes(bs []byte) error
	EncodeByte(b byte) error
	EncodeString(s string) error
	EncodeInt64(i int64) error
}

// Decoder Decodes data types from an underlying io.Reader
type Decoder interface {
	DecodeBytes() ([]byte, error)
	DecodeByte() (byte, error)
	DecodeString() (string, error)
	DecodeInt64() (int64, error)
}

type encoder struct {
	io.Writer
}

type decoder struct {
	io.Reader
}

func NewEncoder(w io.Writer) Encoder {
	return &encoder{
		Writer: w,
	}
}

func NewDecoder(r io.Reader) Decoder {
	return &decoder{
		Reader: r,
	}
}

func (enc *encoder) EncodeBytes(bs []byte) error {
	length := len(bs)
	if length > 254 {
		return fmt.Errorf("wire.EncodeBytes: too long %v", length)
	}
	if _, err := enc.Write([]byte{byte(length)}); err != nil {
		return fmt.Errorf("wire.EncodeBytes: write length: %w", err)
	}
	if _, err := enc.Write(bs); err != nil {
		return fmt.Errorf("wire.EncodeBytes: write payload: %w", err)
	}
	return nil
}

func (enc *encoder) EncodeByte(b byte) error {
	return enc.EncodeBytes([]byte{b})
}

func (enc *encoder) EncodeString(s string) error {
	return enc.EncodeBytes([]byte(s))
}

func (enc *encoder) EncodeInt64(i int64) error {
	bs := &bytes.Buffer{}
	if err := binary.Write(bs, binary.BigEndian, i); err != nil {
		return fmt.Errorf("wire.EncodeInt64: %w", err)
	}
	return enc.EncodeBytes(bs.Bytes())
}

func (dec *decoder) DecodeBytes() ([]byte, error) {
	bs := make([]byte, 1)
	if _, err := dec.Read(bs); err != nil {
		return nil, fmt.Errorf("wire.DecodeBytes length: %w", err)
	}
	length := bs[0]
	if length < 1 {
		return nil, fmt.Errorf("wire.DecodeBytes: bad length: %v", bs[0])
	}

	bs = make([]byte, length)

	_, err := dec.Read(bs)
	if err != nil {
		return nil, fmt.Errorf("wire.DecodeBytes: read payload: %w", err)
	}
	return bs, nil
}

func (dec *decoder) DecodeByte() (byte, error) {
	bs, err := dec.DecodeBytes()
	if err != nil {
		return 0, fmt.Errorf("wire.DecodeByte: %w", err)
	}
	if len(bs) != 1 {
		return 0, fmt.Errorf("wire.DecodeByte: bad length: %v", len(bs))
	}
	return bs[0], nil
}

func (dec *decoder) DecodeString() (string, error) {
	bs, err := dec.DecodeBytes()
	if err != nil {
		return "", fmt.Errorf("wire.DecodeString: %w", err)
	}
	return string(bs), nil
}

func (dec *decoder) DecodeInt64() (int64, error) {
	bs, err := dec.DecodeBytes()
	if err != nil {
		return 0, fmt.Errorf("wire.DecodeInt64: %w", err)
	}

	if len(bs) != sizeOfInt64 {
		return 0, fmt.Errorf("wire.DecodeInt64: bad length: %v", len(bs))
	}

	var i int64
	if err := binary.Read(bytes.NewReader(bs), binary.BigEndian, &i); err != nil {
		return 0, fmt.Errorf("wire.DecodeInt64: %w", err)
	}
	return i, nil
}
