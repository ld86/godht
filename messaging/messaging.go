package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/ld86/godht/types"
)

type IdAddr struct {
	Id   types.NodeID
	Addr string
}

type Message struct {
	FromId     types.NodeID
	ToId       types.NodeID
	Action     string
	Ids        []types.NodeID
	IdsMapping []IdAddr
	IpAddr     *string
}

type Messaging struct {
	serverConnection net.PacketConn
	mapping          map[types.NodeID]net.Addr

	InputMessages  chan Message
	OutputMessages chan Message
}

func (message *Message) String() string {
	return fmt.Sprintf("%s<-%s %s", message.ToId.String(), message.FromId.String(), message.Action)
}

func (messaging *Messaging) GetLocalAddr() string {
	return messaging.serverConnection.LocalAddr().String()
}

func (messaging *Messaging) Serve() {
	go messaging.handleInputMessages()
	messaging.handleOutputMessages()
}

func (messaging *Messaging) handleInputMessages() {
	for {
		var buffer [1024]byte
		var message Message
		n, remoteAddr, _ := messaging.serverConnection.ReadFrom(buffer[:])
		json.Unmarshal(buffer[:n], &message)

		log.Printf("Remember node with id %s by remoteAddr %s", message.FromId.String(), remoteAddr)

		messaging.mapping[message.FromId] = remoteAddr
		for _, idAddr := range message.IdsMapping {
			log.Printf("Remember node with id %v by remoteAddr %s", idAddr.Id, idAddr.Addr)
			messaging.mapping[idAddr.Id], _ = net.ResolveUDPAddr("udp", idAddr.Addr)
		}

		messaging.InputMessages <- message
	}
}

func (messaging *Messaging) handleOutputMessages() {
	for {
		select {
		case outputMessage := <-messaging.OutputMessages:
			var remoteAddr net.Addr

			if outputMessage.IpAddr == nil {
				var ok bool
				remoteAddr, ok = messaging.mapping[outputMessage.ToId]
				if !ok {
					log.Printf("Cannot find remote addr for node with id %s, skipping message", outputMessage.ToId)
					continue
				}
				log.Printf("Found remoteAddr %s by id %s", remoteAddr, outputMessage.ToId.String())
			} else {
				var err error
				remoteAddr, err = net.ResolveUDPAddr("udp", *outputMessage.IpAddr)
				if err != nil {
					log.Printf("Cannot resolve %s, %s", *outputMessage.IpAddr, err)
					continue
				}
				log.Printf("Resolved remoteAddr %s", *outputMessage.IpAddr)
			}

			outputMessage.IdsMapping = make([]IdAddr, 0)
			for _, nodeID := range outputMessage.Ids {
				nodeAddr, ok := messaging.mapping[nodeID]
				if !ok {
					continue
				}
				outputMessage.IdsMapping = append(outputMessage.IdsMapping,
					IdAddr{nodeID, nodeAddr.String()})
			}

			log.Printf("Trying to send message %s", outputMessage.String())
			data, _ := json.Marshal(outputMessage)
			messaging.serverConnection.WriteTo(data, remoteAddr)
		}
	}
}

func createPacketConn() net.PacketConn {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)

	if err != nil {
		log.Fatalf("Cannot create socket, %s", err)
	}

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatalf("Cannot set SO_REUSEADDR on socket, %s", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil && udpAddr.IP != nil {
		log.Fatalf("Cannot resolve addr, %s", err)
	}

	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Port: udpAddr.Port}); err != nil {
		log.Fatalf("Cannot bind socket, %s", err)
	}

	file := os.NewFile(uintptr(fd), string(fd))
	conn, err := net.FilePacketConn(file)
	if err != nil {
		log.Fatalf("Cannot create connection from socket, %s", err)
	}

	if err = file.Close(); err != nil {
		log.Fatalf("Cannot close dup file, %s", err)
	}

	return conn
}

func NewMessaging() *Messaging {
	serverConnection := createPacketConn()
	return &Messaging{serverConnection: serverConnection,
		mapping:        make(map[types.NodeID]net.Addr),
		InputMessages:  make(chan Message),
		OutputMessages: make(chan Message),
	}
}
