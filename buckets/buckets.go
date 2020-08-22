package buckets

import (
	"container/list"
	"errors"
	"math/big"
	"sort"
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

func (buckets *Buckets) innerAddNode(local types.NodeID, remote types.NodeID) (types.NodeID, int, error) {
	buckets.mutex.Lock()
	defer buckets.mutex.Unlock()

	bucketIndex := GetBucketIndex(local, remote)

	if bucketIndex == 0 {
		return local, -1, errors.New("Cannot add yourself to buckets")
	}

	bucketIndex--

	nodeInfo, ok := buckets.nodes[remote]

	if !ok && buckets.buckets[bucketIndex].Len() < buckets.k {
		listPointer := buckets.buckets[bucketIndex].PushBack(remote)
		nodeInfo = &NodeInfo{listPointer: listPointer, UpdateTime: time.Now()}
		buckets.nodes[remote] = nodeInfo
		return remote, bucketIndex, nil
	}

	if ok {
		buckets.buckets[bucketIndex].MoveToBack(nodeInfo.listPointer)
		nodeInfo.UpdateTime = time.Now()
		return remote, bucketIndex, nil
	}

	return buckets.buckets[bucketIndex].Front().Value.(types.NodeID), bucketIndex, errors.New("Please ping this node")
}

func (buckets *Buckets) AddNode(local types.NodeID, remote types.NodeID) (types.NodeID, int, error) {
	responseChan := make(chan interface{})
	request := AddNodeRequest{
		Local:  local,
		Remote: remote,
	}
	message := Message{
		Action:   "AddNode",
		Request:  request,
		Response: responseChan,
	}

	buckets.Messages <- message
	response := (<-responseChan).(AddNodeResponse)
	return response.AddedId, response.BucketIndex, response.Err
}

func (buckets *Buckets) handleAddNode(message Message) {
	request, ok := message.Request.(AddNodeRequest)
	if !ok {
		return
	}
	addedId, bucketIndex, err := buckets.innerAddNode(request.Local, request.Remote)
	response := AddNodeResponse{
		AddedId:     addedId,
		BucketIndex: bucketIndex,
		Err:         err,
	}
	message.Response <- response
}

func (buckets *Buckets) innerRemoveNode(local types.NodeID, remote types.NodeID) (types.NodeID, int, error) {
	buckets.mutex.Lock()
	defer buckets.mutex.Unlock()

	bucketIndex := GetBucketIndex(local, remote)

	if bucketIndex == 0 {
		return local, -1, errors.New("Cannot remove yourself from buckets")
	}

	bucketIndex--
	nodeInfo, ok := buckets.nodes[remote]

	if !ok {
		return remote, bucketIndex, nil
	}

	delete(buckets.nodes, remote)
	buckets.buckets[bucketIndex].Remove(nodeInfo.listPointer)

	return remote, bucketIndex, nil
}

func (buckets *Buckets) RemoveNode(local types.NodeID, remote types.NodeID) (types.NodeID, int, error) {
	responseChan := make(chan interface{})
	request := RemoveNodeRequest{
		Local:  local,
		Remote: remote,
	}
	message := Message{
		Action:   "RemoveNode",
		Request:  request,
		Response: responseChan,
	}

	buckets.Messages <- message
	response := (<-responseChan).(RemoveNodeResponse)
	return response.RemovedId, response.BucketIndex, response.Err
}

func (buckets *Buckets) handleRemoveNode(message Message) {
	request, ok := message.Request.(RemoveNodeRequest)
	if !ok {
		return
	}
	removedId, bucketIndex, err := buckets.innerRemoveNode(request.Local, request.Remote)
	response := RemoveNodeResponse{
		RemovedId:   removedId,
		BucketIndex: bucketIndex,
		Err:         err,
	}
	message.Response <- response
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

type NodeAndDistance struct {
	Id       types.NodeID
	Distance [20]byte
}
type NodesAndDistances []NodeAndDistance

func (nodes NodesAndDistances) Len() int {
	return len(nodes)
}

func (nodes NodesAndDistances) Swap(i, j int) {
	nodes[i], nodes[j] = nodes[j], nodes[i]
}

func (nodes NodesAndDistances) Less(i, j int) bool {
	for index := 0; index < 20; index++ {
		if nodes[i].Distance[index] < nodes[j].Distance[index] {
			return true
		}
	}
	return false
}

func (buckets *Buckets) GetNearestIds(local types.NodeID, remote types.NodeID, k int) []types.NodeID {
	buckets.mutex.Lock()
	defer buckets.mutex.Unlock()

	result := make([]types.NodeID, 0)
	bucketIndex := GetBucketIndex(local, remote)

	if bucketIndex == 0 || buckets.buckets[bucketIndex-1].Len() < k {
		nodesAndDistances := make(NodesAndDistances, 0)
		for nodeId := range buckets.nodes {
			if nodeId == remote {
				continue
			}

			nodeAndDistance := NodeAndDistance{Id: nodeId, Distance: Distance(nodeId, remote)}
			nodesAndDistances = append(nodesAndDistances, nodeAndDistance)
		}

		sort.Sort(nodesAndDistances)

		for i := 0; i < len(nodesAndDistances) && len(result) < k; i++ {
			result = append(result, nodesAndDistances[i].Id)
		}
	} else {
		bucket := buckets.buckets[bucketIndex-1]

		for it := bucket.Back(); it != nil && len(result) < k; it = it.Prev() {
			result = append(result, it.Value.(types.NodeID))
		}
	}

	return result
}
