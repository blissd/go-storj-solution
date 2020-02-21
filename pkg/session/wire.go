package session

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Format a string, such as a file name or a secret code.
// String has max length of 254 bytes.
// Structure is:
// byte 0 -> frame length
// byte 1 -> field name
// byte 2+ -> string
func encodeString(fieldType byte, s string) ([]byte, error) {
	length := uint8(len(s) + 1)
	if length > 255 {
		return nil, errors.New("string too long")
	}

	var buf bytes.Buffer
	buf.WriteByte(length)
	buf.WriteByte(fieldType)
	buf.WriteString(s)
	return buf.Bytes(), nil
}

func decodeString(bs []byte) (byte, string, error) {
	if len(bs) <= 2 {
		return 0, "", errors.New("too short")
	}

	length := int(bs[0])
	if len(bs) != 1+length {
		return 0, "", fmt.Errorf("expect %v bytes, got %v", 2+length, len(bs))
	}

	return bs[1], string(bs[2:]), nil
}

// Encodes a uint32.
// Format is:
// byte 0 -> frame size
// byte 1 -> field type. Always 5.
// byte 2-5 -> big endian encoded length
// No further bytes are encoded, but caller is expected to
// follow this message with exactly `length` bytes.
func encodeUint32(fieldType byte, length uint32) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("msgFileLength must be >= 0")
	}

	var buf bytes.Buffer
	buf.WriteByte(5) // type + sizeof(uint32)
	buf.WriteByte(fieldType)
	if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeUint32(bs []byte) (byte, uint32, error) {
	if len(bs) < 2 {
		return 0, 0, errors.New("too short")
	}

	if bs[0] != 5 {
		return 0, 0, fmt.Errorf("expected length of 5 but got %v", bs[0])
	}

	length := binary.BigEndian.Uint32(bs[2:])
	return bs[1], length, nil
}

// Encodes a single byte. Intended to be used to send flags between processes.
func EncodeByte(b byte) ([]byte, error) {
	return []byte{2, b}, nil
}

func DecodeByte(bs []byte) (byte, error) {
	if len(bs) != 2 {
		return 0, fmt.Errorf("must have length of 3, but is %v bytes", len(bs))
	}

	if bs[0] != 1 {
		return 0, fmt.Errorf("must have encoded length of 1, but is %v bytes", bs[0])
	}

	return bs[1], nil
}
