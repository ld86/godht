package node

import (
    "log"
    "crypto/rand"
    "math/big"
)

const BucketSize = 10

type Node struct {
    id [20]byte
	buckets  *Buckets
}

func NewNodeWithId(id [20]byte) *Node {
	return &Node{id: id, buckets: NewBuckets(BucketSize)}
}

func NewNode() *Node {
    var id [20]byte
    _, err := rand.Read(id[:])
    if err != nil {
        log.Panic("rand.Read failed, %s", err)
    }
    return NewNodeWithId(id)
}
func (node* Node) Distance(secondNode *Node) [20]byte {
    var distance [20]byte
    for i := 0; i < 20; i++ {
        distance[i] = node.id[i] ^ secondNode.id[i]
    }
    return distance
}

func (node* Node) GetBucketIndex(secondNode *Node) int {
    distance := node.Distance(secondNode)

    var intDistance big.Int
    intDistance.SetBytes(distance[:])
    return intDistance.BitLen()
}

func (node* Node) Id() [20]byte {
    return node.id
}


