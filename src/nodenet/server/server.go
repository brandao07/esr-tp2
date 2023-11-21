package server

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/util"
)

func startVideoStreaming(node nodenet.Node, videoFile string) {
	socket := nodenet.SetupSocket("")
	defer socket.Close()

	addr, err := net.ResolveUDPAddr("udp", node.Address+":8080")
	util.HandleError(err)
	videoData, err := os.ReadFile(videoFile)
	util.HandleError(err)

	log.Printf("SERVER: Streaming to %s\n", addr)

	chunks := util.SplitIntoChunks(videoData, 1024)

	i := 0

	for {
		for _, chunk := range chunks {
			nodenet.SendPacket(socket, addr, i, chunk, nodenet.STREAMING)
			time.Sleep(2 * time.Millisecond)
			i++
		}
	}
}

func Run(node *nodenet.Node, videoFile string) {
	go startVideoStreaming(*node, videoFile)
	nodenet.StartServerNode(node, videoFile)
}
