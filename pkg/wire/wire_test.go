package wire

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"unsafe"
)

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{"encode one byte", "a"},
		{"encode two bytes", "ab"},
		{"encode a few words", "some short string"},
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
				t.Fatalf("frame length incorrect. Expected %v, got %v", len(tt.s)+1, len(bs))
			}

			if bs[0] != uint8(len(tt.s)+1) {
				t.Fatalf("Expected encoded frame length to be %v, got %v", len(tt.s)+1, bs[0])
			}

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
		s      string
	}{
		{"decode 'a'", []byte{2, 'a'}, 2, "a"},
		{"decode 'abc'", []byte{4, 'a', 'b', 'c'}, 4, "abc"},
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

			if s != tt.s {
				t.Fatalf("Expected string to be '%v' %v bytes, got '%v' %v bytes", tt.s, len(tt.s), s, len(s))
			}
		})
	}
}

func TestEncodeInt64(t *testing.T) {

	tests := []struct {
		name string
		i    int64
	}{
		{"1", 1},
		{"12", 12},
		{"123", 123},
		{"1234", 1234},
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
		i    int64
	}{
		{"1", 1},
		{"12", 12},
		{"123", 123},
		{"1234", 1234},
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

			if i != tt.i {
				t.Fatalf("Expected i to be %v, got %v", tt.i, i)
			}
		})
	}
}

func Test_frameEncoder_EncodeBytes(t *testing.T) {
	type fields struct {
		Writer bytes.Buffer
	}
	type args struct {
		bs []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"one byte", fields{}, args{[]byte{'a'}}, false},
		{"two bytes", fields{}, args{[]byte{'a', 'b'}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := &encoder{
				Writer: &tt.fields.Writer,
			}
			if err := enc.EncodeBytes(tt.args.bs); (err != nil) != tt.wantErr {
				t.Errorf("EncodeBytes() error = %v, wantErr %v", err, tt.wantErr)
			}

			bs := tt.fields.Writer.Bytes()
			assert.Equal(t, byte(len(tt.args.bs)+1), bs[0])
			assert.True(t, reflect.DeepEqual(tt.args.bs, bs[1:]))
		})
	}
}
