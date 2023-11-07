package server

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/brandao07/esr-tp2/src/util"
)

type Database struct {
	Data map[string]int
	lock sync.RWMutex
}

func service1Func(socket net.PacketConn, sender net.Addr, message string) {
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		fmt.Printf("Received a message from client %s: %s\n", sender.String(), message)
	}

	_, err := socket.WriteTo([]byte("Eu também :)\n"), sender)
	util.HandleError(err)
}

func service2Func(socket net.PacketConn, sender net.Addr, db *Database) {
	(*db).lock.Lock()
	delete((*db).Data, "Lean")
	delete((*db).Data, "Bladee")
	(*db).lock.Unlock()

	_, err := socket.WriteTo([]byte("SUCCESS\n"), sender)
	util.HandleError(err)
}

func service1(wg *sync.WaitGroup) {
	socket, err := net.ListenPacket("udp", os.Args[2])
	util.HandleError(err)

	defer socket.Close()
	defer (*wg).Done()

	fmt.Printf("Listening on %s\n", os.Args[2])
	buffer := make([]byte, 1024)

	for {
		_, sender, err := socket.ReadFrom(buffer)
		util.HandleError(err)

		go service1Func(socket, sender, string(buffer))
	}
}

func service2(wg *sync.WaitGroup, db *Database) {
	socket, err := net.ListenPacket("udp", os.Args[3])
	util.HandleError(err)

	defer socket.Close()
	defer (*wg).Done()

	fmt.Printf("Listening on %s\n", os.Args[3])
	buffer := make([]byte, 1024)

	for {
		_, sender, err := socket.ReadFrom(buffer)
		util.HandleError(err)

		go service2Func(socket, sender, db)
	}
}

func service3(wg *sync.WaitGroup, db *Database) {
	defer (*wg).Done()

	for {
		(*db).lock.Lock()
		for key, value := range ((*db).Data) {
			time.Sleep(2*time.Second)
			fmt.Printf("Key: %s || Init Value: %d || Final Value: %d\n", key, value, (*db).Data[key])
		}
		(*db).lock.Unlock()
	}
}

func Run() {
	var wg sync.WaitGroup
	db := Database {
		Data: make(map[string]int),
	}

	db.Data["Bladee"] = 23
	db.Data["Muse"] = 7
	db.Data["Lean"] = 666

	wg.Add(3)
	go service1(&wg)
	go service2(&wg, &db)
	go service3(&wg, &db)
	wg.Wait()
}
