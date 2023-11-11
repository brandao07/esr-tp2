package example

import (
	"fmt"
	"net"
	"os"

	"github.com/brandao07/esr-tp2/src/util"
)

func RunClient() {
	socket, err := net.ListenPacket("udp", "")
	util.HandleError(err)

	defer socket.Close()

	address, err := net.ResolveUDPAddr("udp", os.Args[3])
	util.HandleError(err)

	_, err = socket.WriteTo([]byte("Adoro Redes"), address)
	util.HandleError(err)

	buffer := make([]byte, 1024)

	_, server, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	fmt.Printf("Received response from server %s: %s\n", server.String(), string(buffer))
}
