package relay

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestEncodeStartSend(t *testing.T) {
	fileName := "some_name.txt"

	bytes, err := EncodeFileName(fileName)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bytes) != 2+len(fileName) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", len(fileName), len(bytes))
	}
	if bytes[0] != msgFileName {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgFileName)
	}
	if bytes[1] != uint8(len(fileName)) {
		t.Fatalf("Expected encoded file name length to be %v, got %v", len(fileName), bytes[1])
	}

	s := string(bytes[2:])
	if s != fileName {
		t.Fatalf("Expected file name to be %v, got %v", fileName, s)
	}
}

func TestDecodeStartSend(t *testing.T) {
	fileName := "some_name.txt"

	bytes, _ := EncodeFileName(fileName)
	s, err := DecodeFileName(bytes)

	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

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
		t.Fatalf("Expected %v, got %v", secret, s)
	}
}

func TestDecodeSecret(t *testing.T) {
	secret := "abc123"

	bytes, _ := EncodeSecret(secret)
	s, err := DecodeSecret(bytes)

	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

	if s != secret {
		t.Fatalf("Expected %v, got %v", secret, s)
	}
}

func TestEncodeContentLength(t *testing.T) {
	length := uint32(231231)

	bytes, err := EncodeFileLength(length)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bytes) != 1+int(unsafe.Sizeof(length)) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", 1+unsafe.Sizeof(length), len(bytes))
	}
	if bytes[0] != msgFileLength {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgFileLength)
	}

	s := binary.BigEndian.Uint32(bytes[1:])
	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}

func TestDecodeContentLength(t *testing.T) {
	length := uint32(231231)

	bytes, _ := EncodeFileLength(length)
	s, err := DecodeFileLength(bytes)
	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}
