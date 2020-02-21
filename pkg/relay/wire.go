package relay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

// Format a messages for initiating the sender side of a transfer session.
// File name cannot length cannot be greater than 255 characters
// Structure is:
// byte 0 -> frame length
// byte 1 -> field name
// byte 2+ -> string
func EncodeString(fieldType byte, s string) ([]byte, error) {
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

func DecodeString(bs []byte) (byte, string, error) {
	if len(bs) <= 2 {
		return 0, "", errors.New("too short")
	}

	length := int(bs[0])
	if len(bs) != 1+length {
		return 0, "", fmt.Errorf("expect %v bytes, got %v", 2+length, len(bs))
	}

	return bs[1], string(bs[2:]), nil
}

// Encodes content length of a file, put to uint32 max value
// Format is:
// byte 0 -> msgFileLength
// byte 1-4 -> big endian encoded length
// No further bytes are encoded, but caller is expected to
// follow this message with exactly `length` bytes.
func EncodeUint32(fieldType byte, length uint32) ([]byte, error) {
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

func DecodeUint32(bs []byte) (byte, uint32, error) {
	if len(bs) < 2 {
		return 0, 0, errors.New("too short")
	}

	if bs[0] != 5 {
		return 0, 0, fmt.Errorf("expected length of 5 but got %v", bs[0])
	}

	length := binary.BigEndian.Uint32(bs[2:])
	return bs[1], length, nil
}
