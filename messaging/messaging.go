package messaging

import (
	"net"
	"log"
	"encoding/json"
)

type Message struct {
	Id [20]byte
	Action string
	Data string
}

type Messaging struct {
	serverConnection *net.UDPConn
	mapping map[[20]byte] *net.UDPAddr

	InputMessages chan Message
	OutputMessages chan Message
}

func (messaging *Messaging) Serve() {
	for {
		var buffer [1024]byte
		var message Message
		n, remoteAddr, _ := messaging.serverConnection.ReadFromUDP(buffer[:])
		json.Unmarshal(buffer[:n], &message)
		messaging.mapping[message.Id] = remoteAddr
		messaging.InputMessages <- message
	}
}

const port = 12334;

func NewMessaging() *Messaging {
	serverSocket, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatalf("Error on binding socket, %s", err)
	}

	serverConnection, err := net.ListenUDP("udp", serverSocket)
	if err != nil {
		log.Fatalf("Error on listening socket, %s", err)
	}

	return &Messaging{serverConnection: serverConnection,
					  InputMessages: make(chan Message),
				      OutputMessages: make(chan Message)}
}
