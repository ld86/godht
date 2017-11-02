package messaging

import (
    "net"
    "log"
    "encoding/json"
)

type Message struct {
    FromId [20]byte
    ToId [20]byte
    Action string
    Data string
}

type Messaging struct {
    serverConnection *net.UDPConn
    serverAddr *net.UDPAddr
    mapping map[[20]byte]string
    bootstrap []string
    id [20]byte

    InputMessages chan Message
    OutputMessages chan Message
}

func (messaging *Messaging) Serve() {
    go messaging.doBootstrap()
    go messaging.handleOutputMessages()
    for {
        var buffer [1024]byte
        var message Message
        n, remoteAddr, _ := messaging.serverConnection.ReadFromUDP(buffer[:])
        json.Unmarshal(buffer[:n], &message)
        messaging.mapping[message.FromId] = remoteAddr.String()
        messaging.InputMessages <- message
    }
}

func (messaging *Messaging) handleOutputMessages() {
    for {
        select {
            case outputMessage := <-messaging.OutputMessages:
                log.Printf("Trying to send message %v", outputMessage)
                remoteIP, ok := messaging.mapping[outputMessage.ToId]
                if !ok {
                    log.Printf("Cannot find remote addr for node with id %s, skipping message")
                    continue
                }
                log.Printf("Found remoteIP %s by id %v", remoteIP, outputMessage.ToId)

                remoteAddr, err := net.ResolveUDPAddr("udp", remoteIP)
                if err != nil {
                    log.Printf("Cannot resolve %s, %s", remoteIP, err)
                    continue
                }

                remoteConn, err := net.DialUDP("udp", messaging.serverAddr, remoteAddr)
                if err != nil {
                    log.Printf("Cannot connect to %s, %s", remoteAddr, err)
                    continue
                }

                defer remoteConn.Close()
                data, _ := json.Marshal(outputMessage)
                log.Printf("%s", remoteConn)
                remoteConn.Write(data)
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

        remoteConn, err := net.DialUDP("udp", messaging.serverAddr, remoteAddr)
        if err != nil {
            log.Printf("Cannot connect to %s, %s", remoteIP, err)
            continue
        }
        defer remoteConn.Close()
        message := Message{FromId: messaging.id, Action: "ping"}
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

    log.Printf("Waiting messages on %s", serverConnection.LocalAddr())

    return &Messaging{serverConnection: serverConnection,
                      serverAddr: serverConnection.LocalAddr().(*net.UDPAddr),
                      mapping: make(map[[20]byte]string),
                      InputMessages: make(chan Message),
                      OutputMessages: make(chan Message),
                      bootstrap: bootstrap,
                      id: id}
}
