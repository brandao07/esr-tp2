package client

import (
	"log"
	"net"
	"sync"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/util"
)

func requestVideoStreaming(socket net.PacketConn, serverAddr, payload string) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte(payload), addr)
	log.Printf("CLIENT: Sending request to %s: %s\n", serverAddr, payload)
	util.HandleError(err)
}

func handleReceivedPacket(buff []byte, addr net.Addr, expectedPacketId uint64) {
	// Write received data to the video file
	pac := nodenet.DecodePacket(buff)
	log.Printf("CLIENT: Received packet from %s: %d bytes\n", addr.String(), len(pac.Data))
	// TODO: feed the video player with the received data

}

func startVideoStreaming(addr, payload string, wg *sync.WaitGroup) {
	socket := nodenet.SetupSocket("")
	defer socket.Close()
	defer (*wg).Done()

	var buff []byte
	var expectedPacketId uint64 = 0

	requestVideoStreaming(socket, addr, payload)

	for {
		buff = make([]byte, 2024)
		_, addr, err := socket.ReadFrom(buff)
		util.HandleError(err)

		go handleReceivedPacket(buff, addr, expectedPacketId)
		expectedPacketId++
	}
}

func Run(serverAddr, videoFile string) {
	var wg sync.WaitGroup

	wg.Add(1)
	go startVideoStreaming(serverAddr, videoFile, &wg)
	wg.Wait()
}
