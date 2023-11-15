package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/brandao07/esr-tp2/src/entity"
	"github.com/brandao07/esr-tp2/src/util"
)

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

// TODO: Implement packet loss detection logic
func isPacketLossDetected(socket net.PacketConn) bool {
	fmt.Println("Verifying for potential packet loss")
	return false
}

func processRequest(socket net.PacketConn, addr net.Addr, request string, videoData []byte) {
	fmt.Printf("Received request from %s: %s\n", addr.String(), request)

	chunks := util.SplitIntoChunks(videoData, 1024)
	for i, chunk := range chunks {
		// Send packet
		util.SendPacket(socket, addr, i, chunk, entity.STREAMING)

		// Check for packet loss and retransmit if necessary
		if isPacketLossDetected(socket) {
			fmt.Printf("Packet loss detected for sequence number: %d\n", i)
			util.SendPacket(socket, addr, i, chunk, entity.STREAMING)
		}
	}

	// Send end of stream signal
	util.SendPacket(socket, addr, 0, nil, entity.FINISHED)
}

func handleUDPRequest(wg *sync.WaitGroup, serverAddress string) {
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

func handleBootstrapRequest(wg *sync.WaitGroup, serverAddress string, isBootstrap bool) {
	defer (*wg).Done()

	if !isBootstrap {
		return
	}

	// Read JSON file
	jsonData, err := os.ReadFile("src/server/bootstrap.json")
	util.HandleError(err)
	
	var data map[string][]entity.Node
	err = json.Unmarshal(jsonData, &data)
	util.HandleError(err)
	fmt.Println(data)

}


func Run(serverAddress string, isBootstrap bool) {
	var wg sync.WaitGroup

	wg.Add(2)
	go handleUDPRequest(&wg, serverAddress)
	go handleBootstrapRequest(&wg, serverAddress, isBootstrap)
	wg.Wait()
}
