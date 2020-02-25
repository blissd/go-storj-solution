package main

import (
	"github.com/blissd/golang-storj-solution/pkg/client"
	"io"
	"log"
	"os"
	"path"
)

func main() {

	if len(os.Args) != 4 {
		log.Fatalln("Usage: receive <relay-host>:<relay-port> <secret-code> <output-directory>")
	}

	addr := os.Args[1]
	secret := os.Args[2]
	dir := os.Args[3]

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		log.Fatalln("output must be an existing directory")
	}

	s, err := client.NewSession(addr)
	if err != nil {
		log.Fatalln("failed creating client:", err)
	}
	defer s.Close()

	if err = s.SendClientTypeReceiver(); err != nil {
		log.Fatalln("failed starting recv client:", err)
	}

	err = s.SendSecret(secret)
	if err != nil {
		log.Fatalln("failed sending secret:", err)
	}

	name, err := s.RecvFileName()
	if err != nil {
		log.Fatalln("failed receiving file name:", err)
	}

	length, err := s.RecvFileLength()
	if err != nil {
		log.Fatalln("failed receiving file length:", err)
	}

	filePath := path.Join(dir, name)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalln("failed creating output file:", err)
	}
	defer file.Close()

	n, err := io.Copy(file, s)
	if err != nil {
		log.Fatalln("failed receiving file:", err)
	}
	if n != length {
		log.Fatalln("incorrect number of bytes received:", n)
	}
}
