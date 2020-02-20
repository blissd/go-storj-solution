package relay

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (

	// Encoded sender type
	msgSend byte = 1

	// Encoded receiver type
	msgRecv byte = 2

	// Encoded secret code
	msgSecretCode byte = 3

	// Encoded file name.
	msgFileName byte = 4

	// Encoded file length
	msgFileLength byte = 5
)

// Format a messages for initiating the sender side of a transfer session.
// File name cannot length cannot be greater than 255 characters
// Structure is:
// byte 0 -> header
// byte 1 -> file name length
// byte 2+ -> file name
func EncodeFileName(fileName string) ([]byte, error) {
	length := uint8(len(fileName))
	if length > 255 {
		return nil, errors.New("file name too long")
	}

	var buf bytes.Buffer
	buf.WriteByte(msgFileName)
	buf.WriteByte(length)
	buf.WriteString(fileName)
	return buf.Bytes(), nil
}

func DecodeFileName(bs []byte) (string, error) {
	if len(bs) <= 2 {
		return "", errors.New("too short")
	}
	if bs[0] != msgFileName {
		return "", fmt.Errorf("expected type %v, got %v", msgFileName, bs[0])
	}
	length := int(bs[1])
	if len(bs) != 2+length {
		return "", fmt.Errorf("expect %v bytes, got %v", 2+length, len(bs))
	}
	return string(bs[2:]), nil
}

// Encodes a secret code, which must be six characters.
// Format is:
// byte 0 -> msgSecretCode
// byte 1-6 -> secret code
func EncodeSecret(secret string) ([]byte, error) {
	if len(secret) != 6 {
		return nil, errors.New("secret code must be six characters")
	}

	var buf bytes.Buffer
	buf.WriteByte(msgSecretCode)
	buf.WriteString(secret)
	return buf.Bytes(), nil
}

func DecodeSecret(bs []byte) (string, error) {
	if len(bs) < 7 || len(bs) > 7 {
		return "", errors.New("secret code must by six characters")
	}

	if bs[0] != msgSecretCode {
		return "", fmt.Errorf("expected type %v, got %v", msgSecretCode, bs[0])
	}
	return string(bs[1:]), nil
}

// Encodes content length of a file, put to uint32 max value
// Format is:
// byte 0 -> msgFileLength
// byte 1-4 -> big endian encoded length
// No further bytes are encoded, but caller is expected to
// follow this message with exactly `length` bytes.
func EncodeFileLength(length uint32) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("msgFileLength must be >= 0")
	}

	var buf bytes.Buffer
	buf.WriteByte(msgFileLength)
	if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeFileLength(bs []byte) (uint32, error) {
	if len(bs) < 5 || len(bs) > 5 {
		return 0, errors.New("must be 5 bytes")
	}

	if bs[0] != msgFileLength {
		return 0, fmt.Errorf("expected type %v, got %v", msgFileLength, bs[0])
	}

	length := binary.BigEndian.Uint32(bs[1:])
	return length, nil
}
