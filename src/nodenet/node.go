package nodenet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/brandao07/esr-tp2/src/util"
)

type Node struct {
	Id         string `json:"id"`
	Address    string `json:"address"`
	Port       string `json:"port"`
	Server     bool   `json:"server"`
	Neighbours []Node `json:"neighbours"`
}

type NodeType string

const (
	DEFAULT NodeType = "DEFAULT"
	RP      NodeType = "RP"
	SERVER  NodeType = "SERVER"
)

var videos = []string{}

var subscribers = []string{}

var publisher = ""

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

		// if the node does not have a publisher yet
		if publisher == "" {
			publisher = addr.String()
			log.Println("NODE: Found a publisher: " + publisher)
		}

		// if the incoming request address is not a different publisher
		if addr.String() != publisher {
			go func() {
				pac := DecodePacket(buffer)
				log.Println("NODE: Already streaming from another publisher")
				pac.State = STOP_STREAMING
				pac.Source = addr.String()
				for _, neighbour := range node.Neighbours {
					sendRequest(*socket, neighbour.Address+":"+neighbour.Port, pac)
				}
			}()
		}
		go streamToSubscribers(buffer, socket)
	}
}

func streamToSubscribers(buffer []byte, socket *net.PacketConn) {
	for _, subscriber := range subscribers {
		go func(buff []byte, sub string) {
			addr, err := net.ResolveUDPAddr("udp", sub)
			util.HandleError(err)
			_, err = (*socket).WriteTo(buffer, addr)
			util.HandleError(err)
		}(buffer, subscriber)
	}
}

func streamingRequestHandler(wg *sync.WaitGroup, node *Node, streamSocket *net.PacketConn, nodeType NodeType) {
	nodeAddr := node.Address + ":" + node.Port
	socket := SetupSocket(nodeAddr)
	log.Printf("NODE: Listening on %s\n", socket.LocalAddr().String())

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
			State: ABORT,
		}
		for _, neighbour := range node.Neighbours {
			sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, &pac)
		}
		os.Exit(0)
	}()

	for {
		buffer = make([]byte, 2024)
		_, addr := ReadFromSocket(socket, buffer)
		req := DecodePacket(buffer)

		if req.State == STOP_STREAMING || req.State == ABORT {
			go func() {
				log.Println("NODE: Stop streaming request received")
				// FIXME: Use full Address instead of Port
				_, reqPort, err := net.SplitHostPort(req.Source)
				util.HandleError(err)
				_, streamPort, err := net.SplitHostPort((*streamSocket).LocalAddr().String())
				util.HandleError(err)
				if (reqPort != streamPort) && req.State != ABORT {
					return
				}
				log.Println("NODE: Stopping streaming to node: " + addr.String())
				subscribers = util.RemoveString(subscribers, addr.String())
				if nodeType != SERVER {
					if len(subscribers) == 0 {
						log.Println("NODE: No more subscribers. Stopping streaming.")
						videos = []string{}
						publisher = ""
						for _, neighbour := range node.Neighbours {
							go sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, req)
						}
					}
				}
			}()
			continue
		}

		// if the incoming request address is not a subscriber
		if !util.ContainsString(subscribers, addr.String()) {
			log.Println("NODE: New subscriber: " + addr.String())
			subscribers = append(subscribers, addr.String())
		}

		// if the node doesnt have the requested video
		if !util.ContainsString(videos, req.File) {
			log.Println("NODE: Video not found. Requesting video: " + req.File)
			videos = append(videos, req.File)
			// if the node is a RP it looks directly for the SERVER
			if nodeType == RP {
				//TODO: ir diretamente aos servidores
				for _, neighbour := range node.Neighbours {
					if neighbour.Server {
						sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, req)
					}
				}
			} else {
				for _, neighbour := range node.Neighbours {
					go sendRequest(*streamSocket, neighbour.Address+":"+neighbour.Port, req)
				}
			}
		}
	}
}

func StartNode(node *Node) {
	var wg sync.WaitGroup
	//FIXME: change to 8080
	socket := SetupSocket("")
	log.Printf("NODE: Streaming on %s\n", socket.LocalAddr().String())

	wg.Add(2)
	go startVideoStreaming(&wg, &socket, node)
	go streamingRequestHandler(&wg, node, &socket, DEFAULT)
	wg.Wait()
}

func StartServerNode(node *Node, videoFile string) {
	var wg sync.WaitGroup
	//FIXME: change to 8080
	socket := SetupSocket("localhost:8080")
	log.Printf("NODE: Streaming on %s\n", socket.LocalAddr().String())
	videos = append(videos, videoFile)

	wg.Add(2)
	go startVideoStreaming(&wg, &socket, node)
	go streamingRequestHandler(&wg, node, &socket, SERVER)
	wg.Wait()
}

func StartRPNode(node *Node) {
	var wg sync.WaitGroup
	//FIXME: change to 8080
	socket := SetupSocket("")
	log.Printf("RP: Streaming on %s\n", socket.LocalAddr().String())

	wg.Add(2)
	go startVideoStreaming(&wg, &socket, node)
	go streamingRequestHandler(&wg, node, &socket, RP)
	wg.Wait()
}
