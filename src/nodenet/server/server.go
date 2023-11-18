package server

import (
	"log"
	"net"
	"os"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/util"
)

func streamVideoChunks(socket net.PacketConn, addr net.Addr, videoData []byte) {
	chunks := util.SplitIntoChunks(videoData, 1024)
	for i, chunk := range chunks {
		nodenet.SendPacket(socket, addr, i, chunk, nodenet.STREAMING)
	}
}

func startVideoStreaming(node nodenet.Node, videoFile string) {
	socket := nodenet.SetupSocket("")
	defer socket.Close()

	addr, err := net.ResolveUDPAddr("udp", node.Address+":"+node.Port)
	util.HandleError(err)

	videoData, err := os.ReadFile(videoFile)
	util.HandleError(err)

	log.Printf("SERVER: Streaming to %s\n", addr)
	for {
		streamVideoChunks(socket, addr, videoData)
	}
}

func Run(node *nodenet.Node, videoFile string) {
	go startVideoStreaming(*node, videoFile)
	nodenet.Run(node)
}
