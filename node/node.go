package node

import (
    "log"
    "crypto/rand"
    "math/big"
)

type Node struct {
    id [20]byte
}

func (node* Node) Distance(secondNode *Node) [20]byte {
    var distance [20]byte
    for i := 0; i < 20; i++ {
        distance[i] = node.id[i] ^ secondNode.id[i]
    }
    return distance
}

func (node* Node) GetBucketIndex(secondNode *Node) uint {
    distance := node.Distance(secondNode)

    var intDistance big.Int
    intDistance.SetBytes(distance[:])
    return uint(intDistance.BitLen())
}

func NewNode() *Node {
    var id [20]byte
    _, err := rand.Read(id[:])
    if err != nil {
        log.Panic("rand.Read failed, %s", err)
    }
    return &Node{id: id}
}
