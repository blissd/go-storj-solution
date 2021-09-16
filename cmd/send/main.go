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

	if err := run(addr, filePath); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(addr string, filePath string) error {

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

	info, err := file.Stat()
	if err != nil {
		log.Fatalln("stating file:", err)
	}

	request := &client.SendRequest{
		Body:   file,
		Name:   file.Name(),
		Length: info.Size(),
	}

	response, err := s.Send(request)
	if err != nil {
		log.Fatalln("sending:", err)
	}

	fmt.Println(response.Secret)

	select {
	case err, gotError := <-response.Errors:
		if gotError {
			log.Fatalln("failed sending file:", err)
		}
	}
	return nil
}
