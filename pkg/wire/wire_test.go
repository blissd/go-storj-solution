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
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			err := enc.EncodeString(tt.s)
			if err != nil {
				t.Fatalf("failed encode: %v", err)
			}

			bs := buf.Bytes()

			if len(bs) != len(tt.s)+1 {
				t.Fatalf("frame length incorrect. Expected %v, got %v", len(tt.s)+2, len(bs))
			}

			if bs[0] != uint8(len(tt.s)+1) {
				t.Fatalf("Expected encoded frame length to be %v, got %v", len(tt.s), bs[0])
			}

			//if bs[1] != tt.id {
			//	t.Fatalf("Expected message type of %v, got %v", bs[0], tt.id)
			//}

			s := string(bs[1:])
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
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.EncodeString(tt.s)

			dec := NewDecoder(&buf)
			s, err := dec.DecodeString()

			if err != nil {
				t.Fatalf("failed decode: %v", err)
			}

			//if id != tt.id {
			//	t.Fatalf("expected type to be %v, got %v", tt.id, id)
			//}

			if s != tt.s {
				t.Fatalf("Expected string to be '%v' %v bytes, got '%v' %v bytes", tt.s, len(tt.s), s, len(s))
			}
		})
	}
}

func TestEncodeInt64(t *testing.T) {

	tests := []struct {
		name string
		id   byte
		i    int64
	}{
		{"1", 34, 1},
		{"12", 2, 12},
		{"123", 127, 123},
		{"1234", 2, 1234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			err := enc.EncodeInt64(tt.i)
			if err != nil {
				t.Fatalf("failed encode: %v", err)
			}
			bs := buf.Bytes()

			if len(bs) != 1+int(unsafe.Sizeof(tt.i)) {
				t.Fatalf("frame length incorrect. Expected %v, got %v", 1+unsafe.Sizeof(tt.i), len(bs))
			}

			if bs[0] != 1+byte(unsafe.Sizeof(tt.i)) {
				t.Fatalf("Expected frame length of %v, got %v", 1+byte(unsafe.Sizeof(tt.i)), bs[0])
			}

			//if bs[1] != tt.id {
			//	t.Fatalf("Expected message type of %v, got %v", bs[0], tt.id)
			//}

			var i int64
			if err := binary.Read(bytes.NewBuffer(bs[1:]), binary.BigEndian, &i); err != nil {
				t.Fatalf("binary read: %v", err)
			}
			if i != tt.i {
				t.Fatalf("Expected i to be %v, got %v", tt.i, i)
			}
		})
	}
}

func TestDecodeInt64(t *testing.T) {
	tests := []struct {
		name string
		id   byte
		i    int64
	}{
		{"1", 34, 1},
		{"12", 2, 12},
		{"123", 127, 123},
		{"1234", 2, 1234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			enc.EncodeInt64(tt.i)

			dec := NewDecoder(&buf)
			i, err := dec.DecodeInt64()

			if err != nil {
				t.Fatalf("failed decode: %v", err)
			}

			//if id != tt.id {
			//	t.Fatalf("expected type of %v, got %v", tt.id, id)
			//}

			if i != tt.i {
				t.Fatalf("Expected i to be %v, got %v", tt.i, i)
			}
		})
	}
}
