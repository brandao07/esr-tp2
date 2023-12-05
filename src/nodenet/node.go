package nodenet

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/brandao07/esr-tp2/src/util"
)

type Node struct {
	Id         string `json:"id"`
	Address    string `json:"address"`
	Port       string `json:"port"`
	ServerPort string `json:"serverPort"`
	IsRP       bool   `json:"isRP"`
	Neighbours []Node `json:"neighbours"`
	Type       NodeType
}

type NodeType string

const (
	DEFAULT NodeType = "DEFAULT"
	RP      NodeType = "RP"
	SERVER  NodeType = "SERVER"
)

type Subscriber struct {
	Id      string
	Address string
}

type Publisher struct {
	Id      string
	Address string
}

type Server struct {
	Id          string
	Address     string
	Latency     float64
	isPublisher bool
}

var switchingServers = false
var lastServer = ""

var videos = []string{}

var servers = []Server{}

var publisher = Publisher{}

var subscribersMutex sync.RWMutex

var subscribers = []Subscriber{}

func findSubscriber(id string) *Subscriber {
	subscribersMutex.RLock()
	defer subscribersMutex.RUnlock()
	for _, sub := range subscribers {
		if sub.Id == id {
			return &sub
		}
	}
	return nil
}

func deleteSubscriber(subscribers *[]Subscriber, sub *Subscriber) error {
	subscribersMutex.Lock()
	defer subscribersMutex.Unlock()
	if subscribers == nil {
		return errors.New("subscribers list is nil")
	}
	for i := range *subscribers {
		if (*subscribers)[i].Id == sub.Id {
			*subscribers = append((*subscribers)[:i], (*subscribers)[i+1:]...)
			return nil
		}
	}
	return errors.New("subscriber not found")
}

func (n *Node) DecodeNode(buff []byte) {
	// Decode node
	decodeBuff := bytes.NewBuffer(buff)
	gobobj := gob.NewDecoder(decodeBuff)
	err := gobobj.Decode(n)
	util.HandleError(err)
}

func (n *Node) EncodeNode() []byte {
	// Encode node
	encodeBuff := new(bytes.Buffer)
	enc := gob.NewEncoder(encodeBuff)
	err := enc.Encode(n)
	util.HandleError(err)

	return encodeBuff.Bytes()
}

func SetupSocket(addr string) net.PacketConn {
	// Listen for UDP packets on the auxiliary address
	socket, err := net.ListenPacket("udp", addr)
	util.HandleError(err)

	return socket
}

func ReadFromSocket(socket net.PacketConn, buffer []byte) (int, net.Addr) {
	n, sender, err := socket.ReadFrom(buffer)
	util.HandleError(err)

	return n, sender
}

func receiveNode(socket net.PacketConn) *Node {
	// Read node
	buff := make([]byte, 2024)
	n, _ := ReadFromSocket(socket, buff)

	// If node not on bootstrap
	if string(buff[:n]) == "NOT_FOUND" {
		return nil
	}

	node := Node{}
	node.DecodeNode(buff[:n])

	return &node
}

func GetNode(bootAddr, id string) *Node {
	// Listen for UDP packets on the auxiliary address
	socket := SetupSocket("")
	defer socket.Close()

	// Resolve the UDP address of the bootstrap server
	addr, err := net.ResolveUDPAddr("udp", bootAddr)
	util.HandleError(err)

	// Send a "GET_NODE" request to the bootstrap server
	_, err = socket.WriteTo([]byte(id), addr)
	util.HandleError(err)

	// Receive the Node information from the bootstrap server
	node := receiveNode(socket)

	if node == nil {
		util.HandleError(fmt.Errorf("node not found"))
	}

	return node
}

func sendRequest(socket net.PacketConn, serverAddr string, pac *Packet) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	util.HandleError(err)
	log.Printf("NODE: Sending request to %s: %s\n", serverAddr, pac.State)
	_, err = socket.WriteTo(EncodePacket(pac), addr)
	util.HandleError(err)
}

func startVideoStreaming(wg *sync.WaitGroup, socket *net.PacketConn, node *Node) {
	defer (*wg).Done()
	defer (*socket).Close()
	var buffer []byte
	for {
		buffer = make([]byte, 2024)
		_, addr := ReadFromSocket(*socket, buffer)
		pac := DecodePacket(buffer)
		// if the node does not have a publisher yet
		if publisher.Id == "" || switchingServers && pac.Source != lastServer {
			if node.Type == RP {
				for i := range servers {
					if servers[i].Id == pac.Source {
						servers[i].isPublisher = true
					}
				}
			}
			lastServer = pac.Source
			publisher.Id = pac.Source
			publisher.Address = addr.String()
			switchingServers = false
			log.Println("NODE: Found a publisher: " + publisher.Id)
		}

		// if the incoming request address is not a different publisher
		if pac.Source != publisher.Id {
			log.Println("NODE: Packet received from another publisher (" + pac.Source + ")")
			log.Println("NODE: Already streaming from another publisher (" + publisher.Id + ")")

			for _, neighbour := range node.Neighbours {
				if neighbour.Id == pac.Source {
					pac := Packet{
						Source: node.Id,
						State:  STOP_STREAMING,
					}
					sendRequest(*socket, neighbour.Address+":"+neighbour.Port, &pac)
					break
				}
			}
			continue
		}
		pac.Source = node.Id
		go streamToSubscribers(EncodePacket(pac), socket, node)
	}
}

func streamToSubscribers(buffer []byte, socket *net.PacketConn, node *Node) {
	if subscribers == nil {
		return
	}

	go func() {
		for _, sub := range subscribers {
			addr, err := net.ResolveUDPAddr("udp", sub.Address)
			util.HandleError(err)
			_, err = (*socket).WriteTo(buffer, addr)
			util.HandleError(err)
		}
	}()
}

func isStopRequest(node *Node, pac *Packet, addr net.Addr, streamSocket *net.PacketConn) bool {
	if pac.State != STOP_STREAMING && pac.State != ABORT {
		return false
	}

	sub := findSubscriber(pac.Source)

	if sub == nil && pac.State == STOP_STREAMING {
		return false
	}

	if pac.State == STOP_STREAMING {
		if pac.Source == node.Id {
			return false
		}
		log.Println("NODE: STOP streaming request received")
		log.Println("NODE: Stopping streaming to node: " + pac.Source)
	}

	if pac.State == ABORT {
		log.Println("NODE: ABORT streaming request received")
		log.Println("NODE: Stopping streaming to node: " + pac.Source)
	}

	if sub == nil && pac.State == ABORT {
		return true
	}

	// delete subscriber from list
	err := deleteSubscriber(&subscribers, sub)
	util.HandleError(err)
	log.Println("NODE: Subscriber " + sub.Id + " disconnected")

	if node.Type == SERVER {
		return true
	}
	if len(subscribers) == 0 {
		log.Println("NODE: No more subscribers. Stopping streaming.")
		videos = []string{}

		pac := Packet{
			Source: node.Id,
			State:  STOP_STREAMING,
		}

		for _, neighbour := range node.Neighbours {
			if neighbour.Id == publisher.Id {
				sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, &pac)
				publisher = Publisher{}
				break
			}
		}
	}

	return true
}

func handleNewServer(node *Node, pac *Packet, addr net.Addr) error {
	if node.Type != RP {
		return errors.New("node is not a RP")
	}
	rpTimestamp := time.Now()
	latency := rpTimestamp.Sub(pac.Timestamp).Seconds()
	server := Server{
		Id:          pac.Source,
		Address:     addr.String(),
		Latency:     latency,
		isPublisher: false,
	}
	log.Println("RP: New Server detected: " + server.Id + " With latency: " + fmt.Sprintf("%f", server.Latency) + "s")
	log.Println(server.Address)
	servers = append(servers, server)
	return nil
}

func checkServers(node *Node, wg *sync.WaitGroup, requestSocket *net.PacketConn, streamSocket *net.PacketConn) {
	socket := SetupSocket("")

	defer socket.Close()
	defer (*wg).Done()

	buffer := make([]byte, 2024)

	for {
		time.Sleep(5 * time.Second)
		if publisher.Id == "" {
			continue
		}
		pac := Packet{
			Source: node.Id,
			State:  CHECK_SERVER,
		}
		if len(servers) <= 1 {
			continue
		}
		log.Println("RP: Checking servers latencies...")
		var currentPublisher Server
		for _, server := range servers {
			sendRequest(socket, server.Address, &pac)
			_, _ = ReadFromSocket(socket, buffer)
			pac := DecodePacket(buffer)
			if pac.State != SERVER_INFO {
				continue
			}
			server.Latency = time.Since(pac.Timestamp).Seconds()
			log.Println("RP: Server " + server.Id + " latency: " + fmt.Sprintf("%f", server.Latency) + "s")
			if server.isPublisher {
				currentPublisher = server
			}
		}

		// compare latencies
		var bestServer *Server

		for i, server := range servers {
			if i == 0 || server.Latency < bestServer.Latency {
				bestServer = &servers[i]
			}
		}

		if bestServer == nil || currentPublisher.Id == "" {
			continue
		}

		// check if the best server is 5x better than the current publisher
		if currentPublisher.Latency/5 >= bestServer.Latency {
			log.Println("RP: Switching to server: " + bestServer.Id)
			pac := Packet{
				Source: node.Id,
				State:  STOP_STREAMING,
			}
			sendRequest(*requestSocket, currentPublisher.Address, &pac)
			publisher.Id = ""
			publisher.Address = ""
			pac.State = REQUESTING
			pac.File = videos[0]
			switchingServers = true
			sendRequest(*streamSocket, bestServer.Address, &pac)
		} else {
			log.Println("RP: Staying with current publisher: " + currentPublisher.Id)
		}
	}
}

func streamingRequestHandler(wg *sync.WaitGroup, node *Node, streamSocket *net.PacketConn) {
	nodeAddr := node.Address + ":" + node.Port
	socket := SetupSocket(nodeAddr)
	log.Printf("NODE: Listening on %s\n", socket.LocalAddr().String())

	if node.Type == SERVER {
		go notifyRP(wg, node, &socket)
	}

	if node.Type == RP {
		wg.Add(1)
		go checkServers(node, wg, &socket, streamSocket)
	}

	defer socket.Close()
	defer (*wg).Done()

	var buffer []byte

	// Handling interrupt signal (CTRL+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		fmt.Println("\nReceived interrupt signal. Cleaning up...")
		pac := Packet{
			Source: node.Id,
			State:  ABORT,
		}
		for _, neighbour := range node.Neighbours {
			sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, &pac)
		}
		os.Exit(0)
	}()

	for {
		buffer = make([]byte, 2024)
		_, addr := ReadFromSocket(socket, buffer)
		pac := DecodePacket(buffer)
		if pac.State == CHECK_SERVER {
			pac.Source = node.Id
			pac.State = SERVER_INFO
			pac.Timestamp = time.Now()
			sendRequest(socket, addr.String(), pac)
			continue
		}
		if pac.State == NEW_SERVER {
			err := handleNewServer(node, pac, addr)
			util.HandleError(err)
			continue
		}
		//fmt.Println("NODE: Received request from " + pac.Source)
		// checks for stop streaming requests
		if isStopRequest(node, pac, addr, streamSocket) {
			continue
		}

		sub := findSubscriber(pac.Source)
		// if the incoming request address is not a subscriber
		if sub == nil {
			if pac.State == STOP_STREAMING || pac.State == ABORT {
				continue
			}

			log.Println("NODE: New subscriber: " + pac.Source)
			newSub := Subscriber{
				Id:      pac.Source,
				Address: addr.String(),
			}
			subscribersMutex.Lock()
			subscribers = append(subscribers, newSub)
			sub = &newSub
			subscribersMutex.Unlock()
		}

		if node.Type == SERVER {
			continue
		}

		// if the node does have the requested video
		if util.ContainsString(videos, pac.File) {
			continue
		}
		log.Println("NODE: Video not found. Requesting video: " + pac.File)
		videos = append(videos, pac.File)
		// if the node is a RP it looks directly for the SERVER
		if node.Type == RP {
			if len(servers) == 0 {
				log.Println("RP: No servers found. Waiting for servers to connect...")
				continue
			}
			for _, server := range servers {
				pac.Source = node.Id
				sendRequest(*streamSocket, server.Address, pac)
			}
		} else {
			for _, neighbour := range node.Neighbours {
				if neighbour.Id == sub.Id {
					continue
				}
				pac.Source = node.Id
				go sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, pac)
			}
		}
	}
}

// Sends a notification to RP that a new server has been detected
func notifyRP(wg *sync.WaitGroup, node *Node, socket *net.PacketConn) {
	defer (*wg).Done()
	if node.Type != SERVER {
		return
	}
	pac := Packet{
		Source:    node.Id,
		State:     NEW_SERVER,
		Timestamp: time.Now(),
	}
	for _, neighbour := range node.Neighbours {
		if neighbour.IsRP {
			log.Println("SERVER: Notifying RP: " + neighbour.Id)
			sendRequest(*socket, neighbour.Address+":"+neighbour.Port, &pac)
		}
	}
}

func StartNode(node *Node) {
	var wg sync.WaitGroup
	socket := SetupSocket("")
	log.Printf("NODE: Streaming on %s\n", socket.LocalAddr().String())

	node.Type = DEFAULT

	wg.Add(2)
	go startVideoStreaming(&wg, &socket, node)
	go streamingRequestHandler(&wg, node, &socket)
	wg.Wait()
}

func StartServerNode(node *Node, videoFile string) {
	var wg sync.WaitGroup
	socket := SetupSocket(node.Address + ":" + node.ServerPort)
	log.Printf("NODE: Streaming on %s\n", socket.LocalAddr().String())
	videos = append(videos, videoFile)

	node.Type = SERVER

	wg.Add(2)
	go startVideoStreaming(&wg, &socket, node)
	go streamingRequestHandler(&wg, node, &socket)
	wg.Wait()
}

func StartRPNode(node *Node) {
	var wg sync.WaitGroup
	socket := SetupSocket("")
	log.Printf("RP: Streaming on %s\n", socket.LocalAddr().String())
	log.Println("RP: Waiting for servers to connect...")
	node.Type = RP

	wg.Add(2)
	go startVideoStreaming(&wg, &socket, node)
	go streamingRequestHandler(&wg, node, &socket)
	wg.Wait()
}
