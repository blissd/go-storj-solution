package relay

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const (
	// Client initiating file send.
	msgSendStart byte = 1

	// Server informing sending client of secret code
	// or receiving client initiating file receipt with secret code.
	msgSecretCode byte = 3

	// Informs sending client they can start sending file data.
	// File data cannot be sent until a receiving client is ready.
	msgReady byte = 4

	// Informs server or receiving client of the file length and data
	msgContents byte = 5
)

// Format a messages for initiating the sender side of a transfer session.
// File name cannot length cannot be greater than 255 characters
// Structure is:
// byte 0 -> header
// byte 1 -> file name length
// byte 2+ -> file name
func EncodeStartSend(fileName string) ([]byte, error) {
	length := uint8(len(fileName))
	if length > 255 {
		return nil, errors.New("File name too long")
	}

	var buf bytes.Buffer
	buf.WriteByte(msgSendStart)
	buf.WriteByte(uint8(length))
	buf.WriteString(fileName)
	return buf.Bytes(), nil
}

// Encodes a secret code, which must be six characters.
// Format is:
// byte 0 -> msgSecretCode
// byte 1-6 -> secret code
func EncodeSecret(secret string) ([]byte, error) {
	if len(secret) != 6 {
		return nil, errors.New("Secret code must be six characters")
	}

	var buf bytes.Buffer
	buf.WriteByte(msgSecretCode)
	buf.WriteString(secret)
	return buf.Bytes(), nil
}

// Encodes content length of a file, put to uint32 max value
// Format is:
// byte 0 -> msgContents
// byte 1-4 -> big endian encoded length
// No further bytes are encoded, but caller is expected to
// follow this message with exactly `length` bytes.
func EncodeContentLength(length uint32) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("msgContents must be >= 0")
	}

	var buf bytes.Buffer
	buf.WriteByte(msgContents)
	if err := binary.Write(&buf, binary.BigEndian, length); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
