package client

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/brandao07/esr-tp2/src/entity"
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
	util.HandleError(err)
}

func handleReceivedPacket(socket net.PacketConn, addr net.Addr, videoFile *os.File, expectedPacketId uint64, pac *entity.Packet) {
	fmt.Printf("Packet #%d received, containing %d bytes of data\n", expectedPacketId, len(pac.Data))
	util.SendPacket(socket, addr, int(expectedPacketId), []byte{}, entity.ACKNOWLEDGE)

	// Write received data to the video file
	_, err := videoFile.Write(pac.Data[:len(pac.Data)])
	util.HandleError(err)
}

// TODO: Server should retransmit packets that are lost
func handleLostPacket(socket net.PacketConn, addr net.Addr, expectedPacketId uint64) {
	fmt.Printf("Packet #%d is missing and presumed lost\n", expectedPacketId)
	util.SendPacket(socket, addr, int(expectedPacketId), []byte{}, entity.REQUESTING)
}

func handleEndOfStream() {
	fmt.Println("Received end of stream signal")
}

func readResponse(socket net.PacketConn, videoFile *os.File) {
	var expectedPacketId uint64 = 0

	for {
		pac, addr := util.ReceivePacket(socket)

		// Check for end of stream signal
		if entity.PacketState(pac.State) == entity.FINISHED {
			handleEndOfStream()
			break
		}

		// Retrieve packet id
		id, err := util.GetPacketId(pac)
		util.HandleError(err)

		// Handle lost, received, or retransmitted packets
		switch {
		case id < expectedPacketId:
			fmt.Printf("Received packet #%d, current expected packet is #%d. Ignoring.\n", id, expectedPacketId)
		case id > expectedPacketId:
			handleLostPacket(socket, addr, expectedPacketId)
		case id == expectedPacketId:
			handleReceivedPacket(socket, addr, videoFile, expectedPacketId, pac)
			expectedPacketId++
		}
	}
}

func Run(serverAddress, request string) {
	socket := setupClient()
	defer socket.Close()

	// Create a file to write the incoming video data
	videoFile, err := os.Create("out.mjpeg")
	util.HandleError(err)
	defer videoFile.Close()

	// Start playing the video file with VLC
	cmd := exec.Command("open", "-a", "vlc", "out.mjpeg")
	err = cmd.Start()
	util.HandleError(err)

	sendRequest(socket, serverAddress, request)
	readResponse(socket, videoFile)
}
