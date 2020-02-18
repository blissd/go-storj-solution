package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatalln("Usage: relay :<port>")
	}

	port := os.Args[1]
	fmt.Println("Port is", port)

}
