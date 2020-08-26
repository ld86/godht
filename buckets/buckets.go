package buckets

import (
	"container/list"
	"math/big"
	"sync"
	"time"

	"github.com/ld86/godht/types"
)

type NodeInfo struct {
	listPointer *list.Element
	UpdateTime  time.Time
}

type AddNodeResponse struct {
	AddedId     types.NodeID
	BucketIndex int
	Err         error
}

type AddNodeRequest struct {
	Local  types.NodeID
	Remote types.NodeID
}

type RemoveNodeResponse struct {
	RemovedId   types.NodeID
	BucketIndex int
	Err         error
}

type RemoveNodeRequest struct {
	Local  types.NodeID
	Remote types.NodeID
}

type Message struct {
	Action   string
	Request  interface{}
	Response chan interface{}
}

type Buckets struct {
	k       int
	buckets [160]list.List
	nodes   map[types.NodeID]*NodeInfo
	mutex   *sync.Mutex

	Messages chan Message
}

func NewBuckets(k int) *Buckets {
	return &Buckets{k: k,
		nodes:    make(map[types.NodeID]*NodeInfo),
		mutex:    &sync.Mutex{},
		Messages: make(chan Message),
	}
}

func (buckets *Buckets) Serve() {
	buckets.handleMessages()
}

func (buckets *Buckets) handleMessages() {
	for {
		select {
		case message := <-buckets.Messages:
			switch message.Action {
			case "AddNode":
				buckets.handleAddNode(message)
			case "RemoveNode":
				buckets.handleRemoveNode(message)
			}
		}
	}
}

func Distance(node types.NodeID, secondNode types.NodeID) [20]byte {
	var distance [20]byte
	for i := 0; i < 20; i++ {
		distance[i] = node[i] ^ secondNode[i]
	}
	return distance
}

func GetBucketIndex(node types.NodeID, secondNode types.NodeID) int {
	distance := Distance(node, secondNode)

	var intDistance big.Int
	intDistance.SetBytes(distance[:])
	return intDistance.BitLen()
}
func (buckets *Buckets) GetBucket(index int) *list.List {
	return &buckets.buckets[index]
}

func (buckets *Buckets) GetSizes() map[int]int {
	result := make(map[int]int)
	for i := 0; i < 160; i++ {
		if l := buckets.buckets[i].Len(); l > 0 {
			result[i] = l
		}
	}
	return result
}

func (buckets *Buckets) GetNodeInfo(nodeId types.NodeID) (*NodeInfo, bool) {
	nodeInfo, found := buckets.nodes[nodeId]
	return nodeInfo, found
}
