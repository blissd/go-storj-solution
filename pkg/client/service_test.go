package client

import (
	"bytes"
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"go-storj-solution/pkg/wire"
	"io"
	"strings"
	"sync"
	"testing"
)

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
		require.NoError(t, err)
		require.Equal(t, "abc", r.Secret)

		sendErr := <-r.Errors
		require.NoError(t, sendErr)
	}()

	// following reads and writes simulate server-side of connection
	bs := []byte{0, 0}
	io.ReadFull(fromClient, bs)
	require.Equal(t, byte('b'), bs[0])
	require.Equal(t, MsgSend, Side(bs[1]))

	toClient.Write([]byte{'s', byte(len(secret))})
	toClient.Write(secret)
	toClient.Write([]byte{'b', byte(MsgRecv)}) // indicates receiver is ready

	// read name
	bs = make([]byte, len(request.Name)+2) // 1 for header 1 for length
	io.ReadFull(fromClient, bs)
	require.Equal(t, byte('s'), bs[0])
	require.Equal(t, byte(len(request.Name)), bs[1])

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
	require.Equal(t, int64(len(body)), length)

	b := &strings.Builder{}
	io.Copy(b, r)
	require.Equal(t, body, b.String())
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
		require.Equal(t, byte('b'), bs[0])
		require.Equal(t, MsgRecv, Side(bs[1]))
		println(2)
		// expect secret
		bs = []byte{0, 0}
		io.ReadFull(fromClient, bs)
		require.Equal(t, byte('s'), bs[0])
		require.Equal(t, len(secret), int(bs[1]))

		bs = make([]byte, bs[1])
		io.ReadFull(fromClient, bs)
		require.Equal(t, []byte(secret), bs)

		// send file name
		toClient.Write([]byte{'s', byte(len(fileName))})
		toClient.Write(fileName)

		// send file body
		toClient.Write([]byte{'B', 0, 0, 0, 0, 0, 0, 0, byte(len(body))})
		toClient.Write(body)
	}()

	s := NewService(wire.NewEncoder(toServer), wire.NewDecoder(fromServer))

	r, err := s.Recv(secret)
	require.NoError(t, err)
	require.Equal(t, string(fileName), r.Name)

	bs := make([]byte, len(body))
	_, err = r.Body.Read(bs)
	require.NoError(t, err)
	require.Equal(t, body, bs)
}
