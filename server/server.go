package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	
	socket, err := net.ListenPacket("udp", os.Args[1])

	if err != nil {
		panic(err.Error())
	}

	defer socket.Close()

	fmt.Printf("Listening on %s\n\n", os.Args[1])
	buffer := make([]byte, 1024)

	_, sender, err := socket.ReadFrom(buffer)

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Received a message from %s: %s\n\n", sender.String(), string(buffer))
}