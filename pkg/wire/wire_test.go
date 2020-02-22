package wire

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestEncodeString(t *testing.T) {
	original := "some text"
	originalId := byte(65)

	bs, err := EncodeString(originalId, original)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}

	if len(bs) != len(original)+2 {
		t.Fatalf("frame length incorrect. Expected %v, got %v", len(original)+1, len(bs))
	}

	if bs[0] != uint8(len(original)+1) {
		t.Fatalf("Expected encoded frame length to be %v, got %v", len(original), bs[0])
	}

	if bs[1] != originalId {
		t.Fatalf("Expected message type of %v, got %v", bs[0], originalId)
	}

	s := string(bs[2:])
	if s != original {
		t.Fatalf("Expected string to be %v, got %v", original, s)
	}
}

func TestDecodeString(t *testing.T) {
	original := "some_name.txt"
	originalId := byte(3)

	bs, _ := EncodeString(originalId, original)
	id, s, err := DecodeString(bs)

	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

	if id != originalId {
		t.Fatalf("expected type to be %v, got %v", originalId, id)
	}

	if s != original {
		t.Fatalf("Expected string to be %v, got %v", original, s)
	}
}

func TestEncodeInt64(t *testing.T) {
	length := int64(231231)
	originalId := byte(8)

	bs, err := EncodeInt64(originalId, length)
	if err != nil {
		t.Fatalf("failed encode: %v", err)
	}
	if len(bs) != 2+int(unsafe.Sizeof(length)) {
		t.Fatalf("frame length incorrect. Expected %v, got %v", 1+unsafe.Sizeof(length), len(bs))
	}

	if bs[0] != 1+byte(unsafe.Sizeof(length)) {
		t.Fatalf("Expected frame length of %v, got %v", 1+byte(unsafe.Sizeof(length)), bs[0])
	}

	if bs[1] != originalId {
		t.Fatalf("Expected message type of %v, got %v", bs[0], originalId)
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
	originalId := byte(127)

	bs, _ := EncodeInt64(originalId, length)
	id, s, err := DecodeInt64(bs)
	if err != nil {
		t.Fatalf("failed decode: %v", err)
	}

	if id != originalId {
		t.Fatalf("expected type of %v, got %v", originalId, id)
	}

	if s != length {
		t.Fatalf("Expected length to be %v, got %v", length, s)
	}
}
