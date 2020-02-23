package main

import (
	"fmt"
	"github.com/blissd/golang-storj-solution/pkg/session"
	"io"
	"log"
	"os"
	"path"
)

func main() {

	if len(os.Args) != 3 {
		log.Fatalln("Usage: send <relay-host>:<relay-port> <file-to-send>")
	}

	addr := os.Args[1]
	filePath := os.Args[2]

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("failed to opening file:", err)
	}
	defer file.Close()

	s, err := session.New(addr)
	if err != nil {
		log.Fatalln("failed creating session:", err)
	}
	defer s.Close()

	if err = s.SendSendReady(); err != nil {
		log.Fatalln("failed starting send session:", err)
	}

	secret, err := s.RecvSecret()
	if err != nil {
		log.Fatalln("failed receiving secret:", err)
	}

	fmt.Println(secret)

	if err = s.WaitForRecv(); err != nil {
		log.Fatalln("failed waiting for receiver:", err)
	}

	if err = s.SendFileName(path.Base(filePath)); err != nil {
		log.Fatalln("failed sending file name:", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		log.Fatalln("failed getting file info:", err)
	}

	if err = s.SendFileLength(info.Size()); err != nil {
		log.Fatalln("failed sending file length:", err)
	}

	n, err := io.Copy(s, file)
	if err != nil {
		log.Fatalln("failed sending file:", err)
	}
	if n != info.Size() {
		log.Fatalln("sent incorrect number of bytes:", n)
	}
}
