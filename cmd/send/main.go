package main

import (
	"fmt"
	"go-storj-solution/pkg/client"
	"log"
	"os"
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

	s, err := client.NewService(addr)
	if err != nil {
		log.Fatalln("failed creating client:", err)
	}
	defer s.Close()

	secret, errs := s.Send(file)
	fmt.Println(secret)

	select {
	case err, gotError := <-errs:
		if gotError {
			log.Fatalln("failed sending file:", err)
		}
	}
}
