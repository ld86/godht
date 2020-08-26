package buckets

import (
	"errors"

	"github.com/ld86/godht/types"
)

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
