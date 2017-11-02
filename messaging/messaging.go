package messaging

import (
    "os"
    "net"
    "log"
    "syscall"
    "encoding/json"
)

type Message struct {
    FromId [20]byte
    ToId [20]byte
    Action string
    Data string
}

type Messaging struct {
    serverConnection net.PacketConn
    mapping map[[20]byte]net.Addr
    bootstrap []string
    id [20]byte

    InputMessages chan Message
    OutputMessages chan Message
}

func (messaging *Messaging) Serve() {
    go messaging.doBootstrap()
    go messaging.handleInputMessages()
    messaging.handleOutputMessages()
}

func (messaging *Messaging) handleInputMessages() {
    for {
        var buffer [1024]byte
        var message Message
        n, remoteAddr, _ := messaging.serverConnection.ReadFrom(buffer[:])
        json.Unmarshal(buffer[:n], &message)

        log.Printf("Remember node with id %v by remoteAddr %s", message.FromId, remoteAddr)

        messaging.mapping[message.FromId] = remoteAddr
        messaging.InputMessages <- message
    }
}

func (messaging *Messaging) handleOutputMessages() {
    for {
        select {
            case outputMessage := <-messaging.OutputMessages:
                log.Printf("Trying to send message %v", outputMessage)
                remoteAddr, ok := messaging.mapping[outputMessage.ToId]
                if !ok {
                    log.Printf("Cannot find remote addr for node with id %s, skipping message")
                    continue
                }
                log.Printf("Found remoteAddr %s by id %v", remoteAddr, outputMessage.ToId)

                data, _ := json.Marshal(outputMessage)
                messaging.serverConnection.WriteTo(data, remoteAddr)
        }
    }
}

func (messaging *Messaging) doBootstrap() {
    for _, remoteIP := range messaging.bootstrap {
        remoteAddr, err := net.ResolveUDPAddr("udp", remoteIP)
        if err != nil {
            log.Printf("Cannot resolve %s, %s", remoteIP, err)
            continue
        }

        message := Message{FromId: messaging.id, Action: "ping"}
        data, _ := json.Marshal(message)
        messaging.serverConnection.WriteTo(data, remoteAddr)
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

func NewMessaging(bootstrap []string, id [20]byte) *Messaging {
    serverConnection := createPacketConn()

    log.Printf("Waiting messages on %s", serverConnection.LocalAddr())

    return &Messaging{serverConnection: serverConnection,
                      mapping: make(map[[20]byte]net.Addr),
                      InputMessages: make(chan Message),
                      OutputMessages: make(chan Message),
                      bootstrap: bootstrap,
                      id: id}
}
