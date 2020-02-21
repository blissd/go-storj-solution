package session

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestEncodeString(t *testing.T) {
	original := "some text"
	fieldType := msgFileName

	bs, err := encodeString(fieldType, original)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}

	if len(bs) != len(original)+2 {
		t.Fatalf("frame length incorrect. Expected %v, got %v", len(original)+1, len(bs))
	}

	if bs[0] != uint8(len(original)+1) {
		t.Fatalf("Expected encoded frame length to be %v, got %v", len(original), bs[0])
	}

	if bs[1] != fieldType {
		t.Fatalf("Expected message type of %v, got %v", bs[0], msgFileName)
	}

	s := string(bs[2:])
	if s != original {
		t.Fatalf("Expected string to be %v, got %v", original, s)
	}
}

func TestDecodeString(t *testing.T) {
	original := "some_name.txt"

	bs, _ := encodeString(msgSecretCode, original)
	fieldType, s, err := decodeString(bs)

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

func TestEncodeInt64(t *testing.T) {
	length := int64(231231)

	bs, err := encodeInt64(msgFileLength, length)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bs) != 2+int(unsafe.Sizeof(length)) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", 1+unsafe.Sizeof(length), len(bs))
	}

	if bs[0] != 5 {
		t.Fatalf("Expected frame length of 5, got %v", bs[0])
	}

	if bs[1] != msgFileLength {
		t.Fatalf("Expected message type of %v, got %v", bs[0], msgFileLength)
	}

	var s int64
	if err := binary.Read(bytes.NewBuffer(bs[2:]), binary.BigEndian, &s); err != nil {
		t.Fatalf("binary read: %v", err)
	}
	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}

func TestDecodeInt64(t *testing.T) {
	length := int64(231231)

	bs, _ := encodeInt64(msgFileLength, length)
	fieldType, s, err := decodeInt64(bs)
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
