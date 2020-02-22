package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

func NextFrame(r io.Reader) ([]byte, error) {
	length := make([]byte, 1)
	_, err := r.Read(length)
	if err != nil {
		return nil, fmt.Errorf("next frame: %w", err)
	}

	frame := make([]byte, length[0]+1)
	frame[0] = length[0]
	_, err = r.Read(frame[1:])
	return frame, err
}

// Format a string, such as a file name or a secret code.
// String has max length of 254 bytes.
// Structure is:
// byte 0 -> frame length
// byte 1 -> field name
// byte 2+ -> string
func EncodeString(id byte, s string) ([]byte, error) {
	length := uint8(len(s) + 1)
	if length > 255 {
		return nil, errors.New("too long")
	}

	var buf bytes.Buffer
	buf.WriteByte(length)
	buf.WriteByte(id)
	buf.WriteString(s)
	return buf.Bytes(), nil
}

func DecodeString(bs []byte) (byte, string, error) {
	if len(bs) <= 2 {
		return 0, "", errors.New("bad frame")
	}

	length := int(bs[0])
	if len(bs) != 1+length {
		return 0, "", fmt.Errorf("expect %v bytes, got %v", 2+length, len(bs))
	}

	return bs[1], string(bs[2:]), nil
}

// Encodes an int64.
// Format is:
// byte 0 -> frame size
// byte 1 -> field type. Always 5.
// byte 2-5 -> big endian encoded length
// No further bytes are encoded, but caller is expected to
// follow this message with exactly `length` bytes.
func EncodeInt64(id byte, length int64) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(1 + byte(unsafe.Sizeof(length)))
	buf.WriteByte(id)
	if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
		return nil, fmt.Errorf("EncodeInt64: %w", err)
	}
	return buf.Bytes(), nil
}

func DecodeInt64(bs []byte) (byte, int64, error) {
	if len(bs) < 2 {
		return 0, 0, errors.New("bad frame")
	}

	var length int64
	if bs[0] != 1+byte(unsafe.Sizeof(length)) {
		return 0, 0, fmt.Errorf("frame too short: %v", bs[0])
	}

	buf := bytes.NewBuffer(bs[2:])
	if err := binary.Read(buf, binary.BigEndian, &length); err != nil {
		return 0, 0, fmt.Errorf("DecodeInt64: %w", err)
	}
	return bs[1], length, nil
}

// Encodes a single byte. Intended to be used to send flags between processes.
func EncodeByte(b byte) ([]byte, error) {
	return []byte{1, b}, nil
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
