package buckets

import (
	"errors"
	"time"

	"github.com/ld86/godht/types"
)

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
