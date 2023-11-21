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

func sendRequest(socket net.PacketConn, serverAddr, payload string) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte(payload), addr)
	log.Printf("CLIENT: Sending request to %s: %s\n", serverAddr, payload)
	util.HandleError(err)
}

func handleReceivedPacket(buff []byte, addr net.Addr, expectedPacketId uint64, videoFile *os.File) {
	pac := nodenet.DecodePacket(buff)
	// Write received data to the video file
	_, err := videoFile.Write(pac.Data[:len(pac.Data)])
	util.HandleError(err)
}

func startVideoStreaming(addr, payload string, wg *sync.WaitGroup, videoFile *os.File) {
	socket := nodenet.SetupSocket("")
	defer socket.Close()
	defer (*wg).Done()

	var buff []byte
	var expectedPacketId uint64 = 0

	sendRequest(socket, addr, payload)

	// Handling interrupt signal (CTRL+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		sendRequest(socket, addr, "STOP_STREAMING")
		os.Exit(0)
	}()

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
	videoFile, err := os.Create(out)
	util.HandleError(err)
	defer videoFile.Close()

	wg.Add(1)
	go startVideoStreaming(serverAddr, filename, &wg, videoFile)
	// $ vlc --no-one-instance out
	// Start playing the video file with VLC after a delay
	time.Sleep(15 * time.Second)
	cmd := exec.Command("open", "-a", "vlc", out)
	err = cmd.Start()
	util.HandleError(err)

	wg.Wait()
}
