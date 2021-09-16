package main

import (
	"go-storj-solution/pkg/client"
	"go-storj-solution/pkg/proxy"
	"go-storj-solution/pkg/wire"
	"log"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalln("Usage: relay :<port>")
	}

	addr := os.Args[1]

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("failed to listen:", err)
	}
	defer l.Close()

	//Secrets := NewFixedSecret("abc123")
	secrets := proxy.NewRandomSecrets(6, time.Now().UnixNano())
	r := newRelay()
	go r.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("failed to accept connection:", err)
		}

		go onboard(r, secrets, conn)
	}
}

// Onboards a new connection for a sender or a receiver.
// For a sender a Secret will be generated and sent to the sender.
// For a receiver a Secret will be read from the connection.
// A valid client then joins a transfer, either creating it for a sender
// or being associated with an existing transform for a receiver.
func onboard(r *relay, secrets proxy.Secrets, conn net.Conn) {

	dec := wire.NewDecoder(conn)
	clientType, err := dec.DecodeByte()
	if err != nil {
		log.Println("failed reading first byte:", err)
		_ = conn.Close()
		return
	}

	log.Println("onboarding for", clientType)

	var secret string

	switch clientType {
	case client.MsgSend:
		log.Println("sending Secret")
		secret = secrets.Secret()
		log.Println("generated Secret is", secret)
		err = wire.NewEncoder(conn).EncodeString(secret)
		if err != nil {
			log.Println("send Secret in onboard:", err)
			_ = conn.Close()
			return
		}
	case client.MsgRecv:
		log.Println("receiving Secret")
		secret, err = dec.DecodeString()
		if err != nil {
			log.Println("recv Secret in onboard:", err)
			_ = conn.Close()
			return
		}
	default:
		log.Println("invalid client type in onboard:", clientType)
		_ = conn.Close()
		return
	}

	c := clientInfo{
		conn:   conn,
		side:   clientType,
		secret: secret,
	}
	r.join(c)
}
