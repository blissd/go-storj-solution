package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/stretchr/testify/assert"
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

	assert.NoError(t, err)
	defer l.Close()

	s, err := NewService(addr)
	assert.NoError(t, err)

	body := "test body"

	request := &SendRequest{
		Body:   strings.NewReader(body),
		Name:   "test.txt",
		Length: int64(len(body)),
	}
	responsec := make(chan *SendResponse)
	go func() {
		r, err := s.Send(request)
		assert.NoError(t, err)
		responsec <- r
	}()

	conn, err := l.Accept()
	assert.NoError(t, err)
	conn.SetDeadline(time.Now().Add(1 * time.Second))

	side := []byte{0, 0}
	io.ReadFull(conn, side)
	assert.Equal(t, byte('b'), side[0])
	assert.Equal(t, MsgSend, Side(side[1]))

	conn.Write([]byte{'s', 3, 'a', 'b', 'c'})
	response := <-responsec
	assert.Equal(t, "abc", response.Secret)

	conn.Write([]byte{'b', byte(MsgRecv)}) // indicates receiver is ready

	// read name
	name := make([]byte, len(request.Name)+2) // 1 for header 1 for length
	io.ReadFull(conn, name)
	assert.Equal(t, byte('s'), name[0])
	assert.Equal(t, byte(len(request.Name)), name[1])

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
	assert.Equal(t, int64(len(body)), length)

	b := &strings.Builder{}
	io.Copy(b, r)
	assert.Equal(t, body, b.String())
}
