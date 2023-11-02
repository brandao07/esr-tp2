package main

import (
	"net"
	"os"
)

func main() {
	socket, err := net.ListenPacket("udp", "")

	if err != nil {
		panic(err.Error())
	}

	defer socket.Close()

	address, err := net.ResolveUDPAddr("udp", os.Args[1])

	if err != nil {
		panic(err.Error())
	}

	_, err = socket.WriteTo([]byte("Adoro Redes"), address)

	if err != nil {
		panic(err.Error())
	}

}