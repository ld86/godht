package node

import (
    "log"
    "crypto/rand"
    "math/big"
	"github.com/ld86/godht/messaging"
)

const BucketSize = 10

type Node struct {
    id [20]byte
	buckets  *Buckets
	messaging *messaging.Messaging
}

func NewNodeWithId(id [20]byte, bootstrap []string) *Node {
	return &Node{id: id,
				 buckets: NewBuckets(BucketSize),
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

func (node *Node) Distance(secondNode *Node) [20]byte {
    var distance [20]byte
    for i := 0; i < 20; i++ {
        distance[i] = node.id[i] ^ secondNode.id[i]
    }
    return distance
}

func (node *Node) GetBucketIndex(secondNode *Node) int {
    distance := node.Distance(secondNode)

    var intDistance big.Int
    intDistance.SetBytes(distance[:])
    return intDistance.BitLen()
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
	log.Println(message)
}
