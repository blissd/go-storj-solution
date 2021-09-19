package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/require"
	"go-storj-solution/pkg/wire"
	"io"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"
)

func Test_service_Send(t *testing.T) {

	port := rand.Intn(1000) + 9000
	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)

	require.NoError(t, err)
	defer l.Close()

	serverCon, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	s := NewService(wire.NewEncoder(serverCon), wire.NewDecoder(serverCon))

	body := "test body"

	request := &SendRequest{
		Body:   strings.NewReader(body),
		Name:   "test.txt",
		Length: int64(len(body)),
	}
	responsec := make(chan *SendResponse)
	go func() {
		r, err := s.Send(request)
		require.NoError(t, err)
		responsec <- r
	}()

	conn, err := l.Accept()
	require.NoError(t, err)
	conn.SetDeadline(time.Now().Add(1 * time.Second))

	side := []byte{0, 0}
	io.ReadFull(conn, side)
	require.Equal(t, byte('b'), side[0])
	require.Equal(t, MsgSend, Side(side[1]))

	conn.Write([]byte{'s', 3, 'a', 'b', 'c'})
	response := <-responsec
	require.Equal(t, "abc", response.Secret)

	conn.Write([]byte{'b', byte(MsgRecv)}) // indicates receiver is ready

	// read name
	name := make([]byte, len(request.Name)+2) // 1 for header 1 for length
	io.ReadFull(conn, name)
	require.Equal(t, byte('s'), name[0])
	require.Equal(t, byte(len(request.Name)), name[1])

	if request.Name != string(name[2:]) {
		t.Fatalf("want %v, got %v", request.Name, string(name[2:]))
	}

	// read body
	bs := make([]byte, len(body)+1+8) // +1 for type +8 for size of int64
	io.ReadFull(conn, bs)

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

	port := rand.Intn(1000) + 9000
	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	require.NoError(t, err)
	defer l.Close()

	serverCon, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	s := NewService(wire.NewEncoder(serverCon), wire.NewDecoder(serverCon))

	secret := "foobar"

	responsec := make(chan *RecvResponse)
	go func() {
		r, err := s.Recv(secret)
		require.NoError(t, err)
		responsec <- r
	}()

	conn, err := l.Accept()
	require.NoError(t, err)
	conn.SetDeadline(time.Now().Add(1 * time.Second))

	side := []byte{0, 0}
	io.ReadFull(conn, side)
	require.Equal(t, byte('b'), side[0])
	require.Equal(t, MsgRecv, Side(side[1]))

	bs := []byte{0, 0}
	io.ReadFull(conn, bs)
	require.Equal(t, byte('s'), bs[0])
	require.Equal(t, len(secret), int(bs[1]))

	fileName := []byte("file.txt")
	conn.Write([]byte{'s', byte(len(fileName))})
	conn.Write(fileName)

	body := []byte("i like cheese")
	conn.Write([]byte{'B', 0, 0, 0, 0, 0, 0, 0, byte(len(body))})
	conn.Write(body)

	response := <-responsec
	require.Equal(t, string(fileName), response.Name)

	bs = make([]byte, len(body))
	_, err = response.Body.Read(bs)
	require.NoError(t, err)
	require.Equal(t, body, bs)
}
