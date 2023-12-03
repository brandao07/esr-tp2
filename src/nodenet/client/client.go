package client

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/brandao07/esr-tp2/src/nodenet"
	"github.com/brandao07/esr-tp2/src/util"
)

func sendRequest(socket net.PacketConn, serverAddr string, pac *nodenet.Packet) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	util.HandleError(err)
	log.Printf("CLIENT: Sending request to %s: %s\n", serverAddr, pac.File)
	_, err = socket.WriteTo(nodenet.EncodePacket(pac), addr)
	util.HandleError(err)
}

func handleReceivedPacket(buff []byte, addr net.Addr, expectedPacketId uint64, videoFile *os.File) {
	pac := nodenet.DecodePacket(buff)
	// Write received data to the video file
	_, err := videoFile.Write(pac.Data[:len(pac.Data)])
	util.HandleError(err)
}

// func checkTimeout(err error) {
// 	netErr, isTimeout := err.(net.Error)
// 	if isTimeout && netErr.Timeout() {
// 		util.HandleError(errors.New("network not responding, please try again later"))
// 		os.Exit(1)
// 	} else if err != nil {
// 		util.HandleError(err)
// 	}
// }

func startVideoStreaming(addr, payload string, wg *sync.WaitGroup, videoFile *os.File, clientId string) {
	socket := nodenet.SetupSocket("")
	defer socket.Close()
	defer (*wg).Done()

	var buff []byte
	var expectedPacketId uint64 = 0

	packet := nodenet.Packet{
		Id:     []byte("0"),
		Source: clientId,
		File:   payload,
		State:  nodenet.REQUESTING,
	}

	sendRequest(socket, addr, &packet)

	// Handling interrupt signal (CTRL+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		pac := nodenet.Packet{
			State:  nodenet.ABORT,
			Source: clientId,
		}

		addr, err := net.ResolveUDPAddr("udp", addr)
		util.HandleError(err)
		log.Printf("CLIENT: Sending request to %s: %s\n", addr, pac.State)
		_, err = socket.WriteTo(nodenet.EncodePacket(&pac), addr)
		util.HandleError(err)
		os.Exit(0)
	}()

	// deadline := time.Now().Add(5 * time.Second)
	// err := socket.SetReadDeadline(deadline)
	// util.HandleError(err)

	for {
		buff = make([]byte, 2024)
		_, addr, err := socket.ReadFrom(buff)
		util.HandleError(err)
		go handleReceivedPacket(buff, addr, expectedPacketId, videoFile)
		expectedPacketId++
	}
}

func Run(serverAddr, filename string) {
	var wg sync.WaitGroup
	// Create a file to write the incoming video data
	out := fmt.Sprintf("./out/out-%d.mjpeg", time.Now().UnixNano()/int64(time.Millisecond))
	clientId := fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
	videoFile, err := os.Create(out)
	util.HandleError(err)
	defer videoFile.Close()

	wg.Add(1)
	go startVideoStreaming(serverAddr, filename, &wg, videoFile, clientId)
	// $ vlc --no-one-instance out
	// Start playing the video file with VLC after a delay
	time.Sleep(1 * time.Second)
	log.Println("CLIENT: Loading stream...")
	time.Sleep(2 * time.Second)
	log.Println("CLIENT: Starting VLC...")
	time.Sleep(2 * time.Second)
	cmd := exec.Command("open", "-a", "vlc", out)
	err = cmd.Start()
	util.HandleError(err)

	wg.Wait()
}
