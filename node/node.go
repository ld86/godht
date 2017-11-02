package node

import (
    "log"
    "time"
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

func (node *Node) pingOldNodes() {
    for {
        for i := 0; i < 160; i++ {
            bucket := node.buckets.GetBucket(i)
            if bucket.Len() > 0 {
                message := messaging.Message{FromId: node.id, ToId: bucket.Front().Value.([20]byte), Action: "ping"}
                node.messaging.OutputMessages <- message
            }
        }
        time.Sleep(60 * time.Second)
    }
}

func (node *Node) Serve() {
    go node.messaging.Serve()
    go node.pingOldNodes()
    for {
        select {
            case message := <-node.messaging.InputMessages:
                node.DispatchMessage(&message)
        }
    }
}

func (node *Node) DispatchMessage(message *messaging.Message) {
    log.Printf("Got message %v", message)
    node.buckets.AddNode(node.id, message.FromId)
    switch message.Action {
        case "ping":
            outputMessage := messaging.Message{FromId: node.id, ToId: message.FromId, Action: "pong"}
            node.messaging.OutputMessages <- outputMessage
    }
}
