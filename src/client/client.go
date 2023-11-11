package client

import (
	"fmt"
	"net"
	"os"
	"os/exec"

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

func readResponse(socket net.PacketConn, videoFile *os.File) {
	buffer := make([]byte, 1024)

	for {
		n, response, err := socket.ReadFrom(buffer)
		util.HandleError(err)

		// Check for end of stream signal
		if string(buffer[:n]) == "END_OF_STREAM" {
			fmt.Println("Received end of stream signal")
			break
		}

		// Write received data to video file
		_, err = videoFile.Write(buffer[:n])
		util.HandleError(err)

		fmt.Printf("Received %d bytes from %s\n", n, response.String())
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
