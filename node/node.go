package node

import (
    "log"
    "time"
    "crypto/rand"
    "github.com/ld86/godht/messaging"
    "github.com/ld86/godht/buckets"
)

const BucketSize = 10

type WaitingTicket struct {
    GotPong bool
}

type Node struct {
    id [20]byte
    buckets  *buckets.Buckets
    messaging *messaging.Messaging

    waiting map[[20]byte]*WaitingTicket
}

func NewNodeWithId(id [20]byte, bootstrap []string) *Node {
    return &Node{id: id,
                 buckets: buckets.NewBuckets(BucketSize),
                 messaging: messaging.NewMessaging(bootstrap, id),
                 waiting: make(map[[20]byte]*WaitingTicket)}
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

func (node *Node) Buckets() *buckets.Buckets {
    return node.buckets
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

func (node *Node) addNodeToBuckets(fromId [20]byte) {
    returnedNodeId, bucketIndex, err := node.buckets.AddNode(node.id, fromId)

    if err == nil {
        log.Printf("Successfuly add remote node to buckets")
        return
    }

    if err != nil && bucketIndex == -1 {
        return
    }

    log.Printf("Bucket %s is full, trying to ping node %v", bucketIndex, returnedNodeId)
    go func() {
        waitingTicket := &WaitingTicket{GotPong: false}
        node.waiting[returnedNodeId] = waitingTicket

        pingMessage := messaging.Message{FromId: node.id, ToId: returnedNodeId, Action: "ping"}
        node.messaging.OutputMessages <- pingMessage

        log.Printf("Waiting 5 seconds for %v", returnedNodeId)
        time.Sleep(5 * time.Second)

        if !waitingTicket.GotPong {
            log.Printf("Did not get pong from %v, removing it", returnedNodeId)
            node.buckets.RemoveNode(node.id, returnedNodeId)
            _, _, err = node.buckets.AddNode(node.id, fromId)
            if err != nil {
                log.Fatalf("Something goes wrong with buckets")
            }
            return
        }
        log.Printf("Got pong from %v, leave it in buckets", returnedNodeId)
    }()
}

func (node *Node) DispatchMessage(message *messaging.Message) {
    log.Printf("Got message %v", message)

    node.addNodeToBuckets(message.FromId)

    switch message.Action {
        case "ping":
            outputMessage := messaging.Message{FromId: node.id, ToId: message.FromId, Action: "pong"}
            node.messaging.OutputMessages <- outputMessage
        case "pong":
            waitingTicket, found := node.waiting[message.FromId]
            if found {
                waitingTicket.GotPong = true;
            }
    }

}
