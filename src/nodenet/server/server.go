package server

import (
	"bytes"
	"encoding/binary"
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
			// Convert packet id to byte array
			idBuff := new(bytes.Buffer)
			err := binary.Write(idBuff, binary.LittleEndian, uint64(i))
			util.HandleError(err)
			pac := nodenet.Packet{
				Id:     idBuff.Bytes(),
				Data:   chunk,
				Source: node.Id,
				State:  nodenet.STREAMING,
			}
			nodenet.SendPacket(socket, addr, &pac)
			time.Sleep(2 * time.Millisecond)
			i++
		}
	}
}

func Run(node *nodenet.Node, videoFile string) {
	go startVideoStreaming(*node, videoFile)
	nodenet.StartServerNode(node, videoFile)
}
