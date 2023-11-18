package server

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/util"
)

func setupServer(addr string) net.PacketConn {
	socket, err := net.ListenPacket("udp", addr)
	util.HandleError(err)

	log.Printf("SERVER: Listening on %s\n", addr)

	return socket
}

func processRequest(socket net.PacketConn, addr net.Addr, request string, videoData []byte) {
	log.Printf("NODE: Received request from %s: %s\n", addr.String(), request)

	chunks := util.SplitIntoChunks(videoData, 1024)
	for i, chunk := range chunks {
		// Send packet
		nodenet.SendPacket(socket, addr, i, chunk, nodenet.STREAMING)
	}
}

func handleUDPRequest(wg *sync.WaitGroup, node *nodenet.Node) {
	socket := setupServer(node.FullAddress)
	buffer := make([]byte, 1024)

	defer socket.Close()
	defer (*wg).Done()

	for {
		n, sender := nodenet.ReadFromSocket(socket, buffer)
		request := string(buffer[:n])

		videoData, err := os.ReadFile(request)
		util.HandleError(err)

		go processRequest(socket, sender, request, videoData)
	}
}

func Run(bootAddr string, id string) {
	var wg sync.WaitGroup

	node := nodenet.GetNode(bootAddr, "")

	wg.Add(1)
	go handleUDPRequest(&wg, node)
	wg.Wait()
}
