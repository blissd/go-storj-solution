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

const byteType = 'b'   // single byte
const streamType = 'B' // arbitrary stream of bytes
const stringType = 's' // short string of up to 256 bytes

// Encoder encodes data types to an underlying io.Writer
type Encoder interface {
	EncodeByte(b byte) error
	EncodeString(s string) error
	EncodeReader(r io.Reader, length int64) error
}

// Decoder Decodes data types from an underlying io.Reader
type Decoder interface {
	DecodeByte() (byte, error)
	DecodeString() (string, error)
	DecodeReader() (io.Reader, error)
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

func (enc *encoder) EncodeByte(b byte) error {
	if _, err := enc.Write([]byte{byte(byteType), b}); err != nil {
		return fmt.Errorf("wire.EncodeByte: write length: %w", err)
	}
	return nil
}

func (enc *encoder) EncodeString(s string) error {
	length := len([]byte(s))
	if length > 255 {
		return fmt.Errorf("wire.EncodeString: too long %v", length)
	}
	bs := bytes.Buffer{}
	bs.WriteByte(byte(stringType))
	bs.WriteByte(byte(length))
	bs.WriteString(s)
	if _, err := enc.Write(bs.Bytes()); err != nil {
		return fmt.Errorf("wire.EncodeString: %w", err)
	}
	return nil
}

func (enc *encoder) EncodeReader(r io.Reader, length int64) error {
	if _, err := enc.Write([]byte{streamType}); err != nil {
		return fmt.Errorf("wire.EncodeReader: %w", err)
	}
	if err := binary.Write(enc, binary.BigEndian, length); err != nil {
		return fmt.Errorf("wire.EncodeReader: %w", err)
	}
	if _, err := io.CopyN(enc, r, length); err != nil {
		return fmt.Errorf("wire.EncodeReader: %w", err)
	}
	return nil
}

func (dec *decoder) DecodeByte() (byte, error) {
	bs := []byte{0, 0}
	_, err := io.ReadFull(dec, bs)
	if err != nil {
		return 0, fmt.Errorf("wire.DecodeByte: %w", err)
	}
	if bs[0] != byteType {
		return 0, fmt.Errorf("wire.DecodeByte: bad type: %v", bs[0])
	}
	return bs[1], nil
}

func (dec *decoder) DecodeString() (string, error) {
	bs := []byte{0, 1}
	_, err := io.ReadFull(dec, bs)
	if err != nil {
		return "", fmt.Errorf("wire.DecodeString: %w", err)
	}
	if bs[0] != stringType {
		return "", fmt.Errorf("wire.DecodeString: bad type: %v", bs[0])
	}
	length := bs[1]
	bs = make([]byte, length)
	_, err = io.ReadFull(dec, bs)
	if err != nil {
		return "", fmt.Errorf("wire.DecodeString: %w", err)
	}
	return string(bs), nil
}

func (dec *decoder) DecodeReader() (io.Reader, error) {
	bs := []byte{0}
	_, err := io.ReadFull(dec, bs)
	if err != nil {
		return nil, fmt.Errorf("wire.DecodeReader: %w", err)
	}
	if bs[0] != streamType {
		return nil, fmt.Errorf("wire.DecodeReader bad type: %v", bs[0])
	}

	var length int64
	if err := binary.Read(dec, binary.BigEndian, &length); err != nil {
		return nil, fmt.Errorf("wire.DecodeReader: %w", err)
	}

	return io.LimitReader(dec, length), nil
}
