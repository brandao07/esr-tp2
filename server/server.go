package server

import (
	"fmt"
	"net"
	"os"
)

func Run() {
	
	socket, err := net.ListenPacket("udp", os.Args[2])

	if err != nil {
		panic(err.Error())
	}

	defer socket.Close()

	fmt.Printf("Listening on %s\n\n", os.Args[2])
	buffer := make([]byte, 1024)

	_, sender, err := socket.ReadFrom(buffer)

	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Received a message from %s: %s\n\n", sender.String(), string(buffer))
}