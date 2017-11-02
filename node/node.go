package node

import (
    "log"
    "crypto/rand"
    "github.com/ld86/godht/messaging"
    "github.com/ld86/godht/buckets"
)

const BucketSize = 10

type Node struct {
    id [20]byte
    buckets  *buckets.Buckets
    messaging *messaging.Messaging
}

func NewNodeWithId(id [20]byte, bootstrap []string) *Node {
    return &Node{id: id,
                 buckets: buckets.NewBuckets(BucketSize),
                 messaging: messaging.NewMessaging(bootstrap, id)}
}

func NewNode(bootstrap []string) *Node {
    var id [20]byte
    _, err := rand.Read(id[:])
    if err != nil {
        log.Panic("rand.Read failed, %s", err)
    }
    return NewNodeWithId(id, bootstrap)
}

func (node *Node) Id() [20]byte {
    return node.id
}

func (node *Node) Serve() {
    go node.messaging.Serve()
    for {
        select {
            case message := <-node.messaging.InputMessages:
                node.DispatchMessage(&message)
        }
    }
}

func (node *Node) DispatchMessage(message *messaging.Message) {
    log.Printf("Got message %v", message)
    switch message.Action {
        case "ping":
            outputMessage := messaging.Message{FromId: node.id, ToId: message.FromId, Action: "pong"}
            node.messaging.OutputMessages <- outputMessage
    }
}
