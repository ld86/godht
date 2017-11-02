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
	serverAddr *net.UDPAddr
	mapping map[[20]byte] *net.UDPAddr
	bootstrap []string
	id [20]byte

	InputMessages chan Message
	OutputMessages chan Message
}

func (messaging *Messaging) Serve() {
	go messaging.doBootstrap()
	for {
		var buffer [1024]byte
		var message Message
		n, remoteAddr, _ := messaging.serverConnection.ReadFromUDP(buffer[:])
		json.Unmarshal(buffer[:n], &message)
		messaging.mapping[message.Id] = remoteAddr
		messaging.InputMessages <- message
	}
}

func (messaging *Messaging) doBootstrap() {
	for _, remoteIP := range messaging.bootstrap {
		remoteAddr, err := net.ResolveUDPAddr("udp", remoteIP)
		if err != nil {
			log.Printf("Cannot resolve %s, %s", remoteIP, err)
			continue
		}

		remoteConn, err := net.DialUDP("udp", messaging.serverAddr, remoteAddr)
		if err != nil {
			log.Printf("Cannot connect to %s, %s", remoteIP, err)
			continue
		}
		defer remoteConn.Close()
		message :=Message{Id: messaging.id, Action:"ping"}
		data, _ := json.Marshal(message)
		remoteConn.Write(data)
	}
}


func NewMessaging(bootstrap []string, id [20]byte) *Messaging {
	serverAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		log.Fatalf("Error on binding socket, %s", err)
	}

	serverConnection, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatalf("Error on listening socket, %s", err)
	}

	return &Messaging{serverConnection: serverConnection,
					  serverAddr: serverAddr,
					  mapping: make(map[[20]byte]*net.UDPAddr),
					  InputMessages: make(chan Message),
				      OutputMessages: make(chan Message),
					  bootstrap: bootstrap,
					  id: id}
}
