package example

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/brandao07/esr-tp2/src/util"
)

type Scenario1Config struct {
	ServerAddress string
	RPAddress     string
}

type Scenario1Database struct {
	Data map[string]string
	lock sync.RWMutex
}

func c1(config Scenario1Config, wg *sync.WaitGroup) {
	fmt.Println("\n\nCLIENT1: Starting Request")
	socket, err := net.ListenPacket("udp", "")
	util.HandleError(err)

	defer socket.Close()

	address, err := net.ResolveUDPAddr("udp", config.RPAddress)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte("Andre"), address)
	util.HandleError(err)

	buffer := make([]byte, 1024)

	_, server, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	fmt.Printf("CLIENT1: Received response from %s: %s\n", server.String(), string(buffer))
}

func c2(config Scenario1Config, wg *sync.WaitGroup) {
	time.Sleep(2 * time.Second)
	fmt.Println("\n\nCLIENT2: Starting Request")
	socket, err := net.ListenPacket("udp", "")
	util.HandleError(err)

	defer socket.Close()

	address, err := net.ResolveUDPAddr("udp", config.RPAddress)
	util.HandleError(err)

	_, err = socket.WriteTo([]byte("Andre"), address)
	util.HandleError(err)

	buffer := make([]byte, 1024)

	_, server, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	fmt.Printf("CLIENT2: Received response from %s: %s\n", server.String(), string(buffer))
}

func rpHandler(socket net.PacketConn, sender net.Addr, request, serverAddress string, db *Scenario1Database) {
	fmt.Printf("RP: Received request from %s: %s\n", sender.String(), request)

	if _, ok := db.Data[request]; ok {
		// RP returns the value directly
		value := db.Data[request]
		_, err := socket.WriteTo([]byte(value), sender)
		util.HandleError(err)
	} else {
		// RP asks the Server for the value and then returns it to the Client
		clientSocket, err := net.ListenPacket("udp", "")
		util.HandleError(err)

		defer clientSocket.Close()

		address, err := net.ResolveUDPAddr("udp", serverAddress)
		util.HandleError(err)

		_, err = clientSocket.WriteTo([]byte(request), address)
		util.HandleError(err)

		buffer := make([]byte, 1024)

		_, server, err := clientSocket.ReadFrom(buffer)
		util.HandleError(err)
		fmt.Printf("RP: Received response from %s: %s\n", server.String(), string(buffer))
		(*db).Data[request] = string(buffer)
		_, err = socket.WriteTo(buffer, sender)
		util.HandleError(err)
	}
}

func rp(config Scenario1Config, wg *sync.WaitGroup) {
	db := Scenario1Database{
		Data: make(map[string]string),
	}
	socket, err := net.ListenPacket("udp", config.RPAddress)
	util.HandleError(err)

	defer socket.Close()
	defer (*wg).Done()

	fmt.Printf("RP: Listening on %s\n", config.RPAddress)

	buffer := make([]byte, 1024)
	for {
		n, sender, err := socket.ReadFrom(buffer)
		util.HandleError(err)
		go rpHandler(socket, sender, string(buffer[:n]), config.ServerAddress, &db)
	}
}

func serverHandler(socket net.PacketConn, sender net.Addr, request, serverAddress string, db *Scenario1Database) {
	fmt.Printf("SERVER: Received request from %s: %s\n", sender.String(), request)

	if _, ok := db.Data[request]; ok {
		// Server returns the value directly
		value := db.Data[request]
		_, err := socket.WriteTo([]byte(value), sender)
		util.HandleError(err)
	} else {
		util.HandleError(errors.New("key not found"))
	}
}

func server(config Scenario1Config, wg *sync.WaitGroup) {
	db := Scenario1Database{
		Data: make(map[string]string),
	}
	db.lock.Lock()
	db.Data["David"] = "Muse"
	db.Data["Andre"] = "Bladee"
	db.lock.Unlock()

	socket, err := net.ListenPacket("udp", config.ServerAddress)
	util.HandleError(err)

	defer socket.Close()
	defer (*wg).Done()

	fmt.Printf("SERVER: Listening on %s\n", config.ServerAddress)

	buffer := make([]byte, 1024)
	for {
		n, sender, err := socket.ReadFrom(buffer)
		util.HandleError(err)
		go serverHandler(socket, sender, string(buffer[:n]), config.ServerAddress, &db)
	}
}

func Scenario1(config Scenario1Config) {
	var wg sync.WaitGroup
	wg.Add(4)
	go server(config, &wg)
	go c1(config, &wg)
	go c2(config, &wg)
	go rp(config, &wg)
	wg.Wait()
}
