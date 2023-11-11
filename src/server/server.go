package server

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/brandao07/esr-tp2/src/util"
)

type Database struct {
	Data map[string]string
}

func setupServer(serverAddress string) net.PacketConn {
	socket, err := net.ListenPacket("udp", serverAddress)
	util.HandleError(err)

	fmt.Printf("Listening on %s\n", serverAddress)

	return socket
}

func readFromSocket(socket net.PacketConn, buffer []byte) (int, net.Addr) {
	n, sender, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	return n, sender
}

func processRequest(socket net.PacketConn, sender net.Addr, request string, videoData []byte) {
	fmt.Printf("SERVER: Received request from %s: %s\n", sender.String(), request)

	chunks := util.SplitIntoChunks(videoData, 1024)

	for _, chunk := range chunks {
		_, err := socket.WriteTo(chunk, sender)
		util.HandleError(err)
	}

	// Send end of stream signal
	_, err := socket.WriteTo([]byte("END_OF_STREAM"), sender)
	util.HandleError(err)
}

func handleUDPRequest(wg *sync.WaitGroup, serverAddress string, db *Database) {
	socket := setupServer(serverAddress)
	buffer := make([]byte, 1024)

	defer socket.Close()
	defer (*wg).Done()

	for {
		n, sender := readFromSocket(socket, buffer)
		request := string(buffer[:n])

		videoData, err := os.ReadFile(request)
		util.HandleError(err)

		go processRequest(socket, sender, request, videoData)
	}
}

func Run(serverAddress string) {
	var wg sync.WaitGroup

	db := Database{
		Data: make(map[string]string),
	}

	db.Data["David"] = "Muse"
	db.Data["Andre"] = "Bladee"

	wg.Add(1)
	go handleUDPRequest(&wg, serverAddress, &db)
	wg.Wait()
}
