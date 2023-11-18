package client

import (
	"log"
	"net"
	"sync"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/util"
)

func setupClient() net.PacketConn {
	socket, err := net.ListenPacket("udp", "")
	util.HandleError(err)
	return socket
}

func sendRequest(socket net.PacketConn, serverAddress, request string) {
	address, err := net.ResolveUDPAddr("udp", serverAddress)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte(request), address)
	log.Printf("CLIENT: Sending request to %s: %s\n", serverAddress, request)
	util.HandleError(err)
}

func handleReceivedPacket(buff []byte, addr net.Addr, expectedPacketId uint64) {
	// Write received data to the video file
	pac := nodenet.DecodePacket(buff)
	
	log.Printf("CLIENT: Received packet from %s: %d bytes\n", addr.String(), len(pac.Data))
	// TODO: feed the video player with the received data
}

func readResponse(socket net.PacketConn, wg *sync.WaitGroup) {
	var expectedPacketId uint64 = 0
	defer (*wg).Done()
	var buff []byte
	for {
		buff = make([]byte, 2024)
		_, addr, err := socket.ReadFrom(buff)
		util.HandleError(err)

		go handleReceivedPacket(buff, addr, expectedPacketId)
		expectedPacketId++
	}
}

func Run(serverAddress, request string) {
	socket := setupClient()
	defer socket.Close()

	var wg sync.WaitGroup

	sendRequest(socket, serverAddress, request)
	wg.Add(1)
	go readResponse(socket, &wg)
	wg.Wait()
}
