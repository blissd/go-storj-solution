package wire

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name string
		id   byte
		s    string
	}{
		{"encode one byte", 100, "a"},
		{"encode two bytes", 101, "ab"},
		{"encode a few words", 101, "some short string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := EncodeString(tt.id, tt.s)
			if err != nil {
				t.Fatalf("failed encode: %v", err)
			}

			if len(bs) != len(tt.s)+2 {
				t.Fatalf("frame length incorrect. Expected %v, got %v", len(tt.s)+2, len(bs))
			}

			if bs[0] != uint8(len(tt.s)+1) {
				t.Fatalf("Expected encoded frame length to be %v, got %v", len(tt.s), bs[0])
			}

			if bs[1] != tt.id {
				t.Fatalf("Expected message type of %v, got %v", bs[0], tt.id)
			}

			s := string(bs[2:])
			if s != tt.s {
				t.Fatalf("Expected string to be %v, got %v", tt.s, s)
			}
		})
	}
}

func TestDecodeString(t *testing.T) {

	tests := []struct {
		name   string
		bs     []byte
		length byte
		id     byte
		s      string
	}{
		{"decode 'a'", []byte{2, 1, 'a'}, 2, 1, "a"},
		{"decode 'abc'", []byte{4, 9, 'a', 'b', 'c'}, 4, 9, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, s, err := DecodeString(tt.bs)

			if err != nil {
				t.Fatalf("failed decode: %v", err)
			}

			if id != tt.id {
				t.Fatalf("expected type to be %v, got %v", tt.id, id)
			}

			if s != tt.s {
				t.Fatalf("Expected string to be %v, got %v", tt.s, s)
			}
		})
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
