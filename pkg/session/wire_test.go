package session

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestEncodeString(t *testing.T) {
	original := "some text"
	fieldType := msgFileName

	bytes, err := encodeString(fieldType, original)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}

	if len(bytes) != len(original)+2 {
		t.Fatalf("frame length incorrect. Expected %v, got %v", len(original)+1, len(bytes))
	}

	if bytes[0] != uint8(len(original)+1) {
		t.Fatalf("Expected encoded frame length to be %v, got %v", len(original), bytes[0])
	}

	if bytes[1] != fieldType {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgFileName)
	}

	s := string(bytes[2:])
	if s != original {
		t.Fatalf("Expected string to be %v, got %v", original, s)
	}
}

func TestDecodeString(t *testing.T) {
	original := "some_name.txt"

	bytes, _ := encodeString(msgSecretCode, original)
	fieldType, s, err := decodeString(bytes)

	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

	if fieldType != msgSecretCode {
		t.Fatalf("expected type to be %v, got %v", msgSecretCode, fieldType)
	}

	if s != original {
		t.Fatalf("Expected string to be %v, got %v", original, s)
	}
}

func TestEncodeUint32(t *testing.T) {
	length := uint32(231231)

	bytes, err := encodeUint32(msgFileLength, length)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bytes) != 2+int(unsafe.Sizeof(length)) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", 1+unsafe.Sizeof(length), len(bytes))
	}

	if bytes[0] != 5 {
		t.Fatalf("Expected frame length of 5, got %v", bytes[0])
	}

	if bytes[1] != msgFileLength {
		t.Fatalf("Expected message type of %v, got %v", bytes[0], msgFileLength)
	}

	s := binary.BigEndian.Uint32(bytes[2:])
	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}

func TestDecodeUint32(t *testing.T) {
	length := uint32(231231)

	bytes, _ := encodeUint32(msgFileLength, length)
	fieldType, s, err := decodeUint32(bytes)
	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

	if fieldType != msgFileLength {
		t.Fatalf("expected type of %v, got %v", msgFileLength, fieldType)
	}

	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}
