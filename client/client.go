package client

import (
	"net"
	"os"
)

func Run() {
	socket, err := net.ListenPacket("udp", "")

	if err != nil {
		panic(err.Error())
	}

	defer socket.Close()

	address, err := net.ResolveUDPAddr("udp", os.Args[2])

	if err != nil {
		panic(err.Error())
	}

	_, err = socket.WriteTo([]byte("Adoro Redes"), address)

	if err != nil {
		panic(err.Error())
	}

}