package wire

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		bs   []byte
	}{
		{"encode one byte", "a", []byte{'s', 1, 'a'}},
		{"encode two bytes", "ab", []byte{'s', 2, 'a', 'b'}},
		{"encode a few words", "a b c", []byte{'s', 5, 'a', ' ', 'b', ' ', 'c'}},
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
			if !reflect.DeepEqual(tt.bs, bs) {
				t.Fatalf("wanted %v, got %v", tt.bs, bs)
			}
		})
	}
}

func TestDecodeString(t *testing.T) {

	tests := []struct {
		name string
		bs   []byte
		s    string
	}{
		{"decode 'a'", []byte{'s', 1, 'a'}, "a"},
		{"decode 'abc'", []byte{'s', 3, 'a', 'b', 'c'}, "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewDecoder(bytes.NewReader(tt.bs))
			s, err := enc.DecodeString()

			if err != nil {
				t.Fatalf("failed decode: %v", err)
			}

			if tt.s != s {
				t.Fatalf("want %v, got %v", tt.s, s)
			}
		})
	}
}

func TestEncodeReader(t *testing.T) {
	tests := []struct {
		name string
		s    string
		bs   []byte
	}{
		{"encode one byte", "a", []byte{'B', 0, 0, 0, 0, 0, 0, 0, 1, 'a'}},
		{"encode two bytes", "ab", []byte{'B', 0, 0, 0, 0, 0, 0, 0, 2, 'a', 'b'}},
		{"encode a few words", "a b c", []byte{'B', 0, 0, 0, 0, 0, 0, 0, 5, 'a', ' ', 'b', ' ', 'c'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			err := enc.EncodeReader(strings.NewReader(tt.s), int64(len(tt.s)))
			if err != nil {
				t.Fatalf("failed encode: %v", err)
			}

			bs := buf.Bytes()
			if !reflect.DeepEqual(tt.bs, bs) {
				t.Fatalf("wanted %v, got %v", tt.bs, bs)
			}
		})
	}
}

func TestDecodeReader(t *testing.T) {
	tests := []struct {
		name string
		bs   []byte
		s    string
	}{
		{"encode one byte", []byte{'B', 0, 0, 0, 0, 0, 0, 0, 1, 'a'}, "a"},
		{"encode two bytes", []byte{'B', 0, 0, 0, 0, 0, 0, 0, 2, 'a', 'b'}, "ab"},
		{"encode a few words", []byte{'B', 0, 0, 0, 0, 0, 0, 0, 5, 'a', ' ', 'b', ' ', 'c'}, "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewDecoder(bytes.NewReader(tt.bs))
			r, err := enc.DecodeReader()
			if err != nil {
				t.Fatalf("err: %v", err)
			}

			b := &strings.Builder{}
			io.Copy(b, r)

			if tt.s != b.String() {
				t.Fatalf("wanted %v, got %v", tt.bs, b.String())
			}
		})
	}
}
