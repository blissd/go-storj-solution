package client

import (
	"bytes"
	"encoding/binary"
	"go-storj-solution/pkg/wire"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func IsEqual(t *testing.T, want interface{}, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("Values not equal. want: %v, got: %v", want, got)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func Test_service_Send(t *testing.T) {
	fromClient, toServer := io.Pipe()
	fromServer, toClient := io.Pipe()

	body := "test body"

	request := &SendRequest{
		Body:   strings.NewReader(body),
		Name:   "test.txt",
		Length: int64(len(body)),
	}

	secret := []byte("abc")

	wg := sync.WaitGroup{}
	wg.Add(1)
	// go routine is the sending client
	go func() {
		s := NewService(wire.NewEncoder(toServer), wire.NewDecoder(fromServer))
		r, err := s.Send(request)
		NoError(t, err)
		IsEqual(t, "abc", r.Secret)

		sendErr := <-r.Errors
		NoError(t, sendErr)
	}()

	// following reads and writes simulate server-side of connection
	bs := []byte{0, 0}
	io.ReadFull(fromClient, bs)
	IsEqual(t, byte('b'), bs[0])
	IsEqual(t, MsgSend, Side(bs[1]))

	toClient.Write([]byte{'s', byte(len(secret))})
	toClient.Write(secret)
	toClient.Write([]byte{'b', byte(MsgRecv)}) // indicates receiver is ready

	// read name
	bs = make([]byte, len(request.Name)+2) // 1 for header 1 for length
	io.ReadFull(fromClient, bs)
	IsEqual(t, byte('s'), bs[0])
	IsEqual(t, byte(len(request.Name)), bs[1])

	if request.Name != string(bs[2:]) {
		t.Fatalf("want %v, got %v", request.Name, string(bs[2:]))
	}

	// read body
	bs = make([]byte, len(body)+1+8) // +1 for type +8 for size of int64
	io.ReadFull(fromClient, bs)

	r := bytes.NewReader(bs)
	if b, _ := r.ReadByte(); b != byte('B') {
		t.Fatalf("want %v, got %v", 'B', b)
	}
	var length int64
	binary.Read(r, binary.BigEndian, &length)
	IsEqual(t, int64(len(body)), length)

	b := &strings.Builder{}
	io.Copy(b, r)
	IsEqual(t, body, b.String())
}

func Test_service_Recv(t *testing.T) {
	fromClient, toServer := io.Pipe()
	fromServer, toClient := io.Pipe()

	secret := "foobar"
	fileName := []byte("file.txt")
	body := []byte("i like cheese")

	// go routine is the server
	go func() {
		// expect client-side indicator
		bs := []byte{0, 0}
		io.ReadFull(fromClient, bs)
		IsEqual(t, byte('b'), bs[0])
		IsEqual(t, MsgRecv, Side(bs[1]))
		println(2)
		// expect secret
		bs = []byte{0, 0}
		io.ReadFull(fromClient, bs)
		IsEqual(t, byte('s'), bs[0])
		IsEqual(t, len(secret), int(bs[1]))

		bs = make([]byte, bs[1])
		io.ReadFull(fromClient, bs)
		IsEqual(t, []byte(secret), bs)

		// send file name
		toClient.Write([]byte{'s', byte(len(fileName))})
		toClient.Write(fileName)

		// send file body
		toClient.Write([]byte{'B', 0, 0, 0, 0, 0, 0, 0, byte(len(body))})
		toClient.Write(body)
	}()

	s := NewService(wire.NewEncoder(toServer), wire.NewDecoder(fromServer))

	r, err := s.Recv(secret)
	NoError(t, err)
	IsEqual(t, string(fileName), r.Name)

	bs := make([]byte, len(body))
	_, err = r.Body.Read(bs)
	NoError(t, err)
	IsEqual(t, body, bs)
}
