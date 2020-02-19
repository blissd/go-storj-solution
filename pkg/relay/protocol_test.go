package relay

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestEncodeStartSend(t *testing.T) {
	fileName := "some_name.txt"

	bytes, err := EncodeStartSend(fileName)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bytes) != 2+len(fileName) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", len(fileName), len(bytes))
	}
	if bytes[0] != msgSendStart {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgSendStart)
	}
	if uint8(bytes[1]) != uint8(len(fileName)) {
		t.Fatalf("Expected encoded file name length to be %v, got %v", len(fileName), uint8(bytes[1]))
	}

	s := string(bytes[2:])
	if s != fileName {
		t.Fatalf("Expected file name to be %v, got %v", fileName, s)
	}
}

func TestEncodeSecret(t *testing.T) {
	secret := "abc123"

	bytes, err := EncodeSecret(secret)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bytes) != 1+len(secret) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", len(secret), len(bytes))
	}
	if bytes[0] != msgSecretCode {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgSecretCode)
	}

	s := string(bytes[1:])
	if s != secret {
		t.Fatalf("Expected secret to be %v, got %v", secret, s)
	}
}

func TestEncodeContentLength(t *testing.T) {
	length := uint32(231231)

	bytes, err := EncodeContentLength(length)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bytes) != 1+int(unsafe.Sizeof(length)) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", 1+unsafe.Sizeof(length), len(bytes))
	}
	if bytes[0] != msgContents {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgContents)
	}

	s := binary.BigEndian.Uint32(bytes[1:])
	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}
