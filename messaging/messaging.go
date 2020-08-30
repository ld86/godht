package messaging

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"

	"github.com/ld86/godht/types"
)

type IdAddr struct {
	Id   types.NodeID
	Addr string
}

type Message struct {
	FromId          types.NodeID
	ToId            types.NodeID
	Action          string
	Ids             []types.NodeID
	IdToAddrMapping []IdAddr
	IpAddr          *string
	TransactionID   *types.TransactionID
	Payload         []byte
}

type Messaging struct {
	serverConnection net.PacketConn

	Mapping              *sync.Map
	transactionReceivers *sync.Map

	receiver       chan Message
	messagesToSend chan Message
}

func NewMessaging() *Messaging {
	serverConnection := createPacketConn()
	return &Messaging{serverConnection: serverConnection,
		Mapping:              &sync.Map{},
		transactionReceivers: &sync.Map{},
		messagesToSend:       make(chan Message, 100),
	}
}

func (messaging *Messaging) SendMessage(message Message) {
	messaging.messagesToSend <- message
}

func (messaging *Messaging) Receiver() chan Message {
	return messaging.receiver
}

func (message *Message) String() string {
	strTransactionID := "nil"
	if message.TransactionID != nil {
		strTransactionID = message.TransactionID.String()
	}
	return fmt.Sprintf("%s<-%s %s %s %s", message.ToId.String(), message.FromId.String(), message.Action, strTransactionID, message.Payload)
}

func (messaging *Messaging) GetLocalAddr() string {
	return messaging.serverConnection.LocalAddr().String()
}

func (messaging *Messaging) SetReceiver(receiver chan Message) {
	messaging.receiver = receiver
}

func (messaging *Messaging) Serve() {
	go messaging.handleReceivedMessages()
	messaging.handleSentMessages()
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
