package buckets

import (
    "sync"
    "time"
    "math/big"
    "errors"
    "container/list"
)

type NodeInfo struct {
    listPointer *list.Element
    UpdateTime time.Time
}

type Buckets struct {
    k int
    buckets [160]list.List
    nodes map[[20]byte]*NodeInfo
    mutex *sync.Mutex
}

func NewBuckets(k int) *Buckets {
    return &Buckets{k: k,
                    nodes: make(map[[20]byte]*NodeInfo),
                    mutex: &sync.Mutex{}}
}

func Distance(node [20]byte, secondNode [20]byte) [20]byte {
    var distance [20]byte
    for i := 0; i < 20; i++ {
        distance[i] = node[i] ^ secondNode[i]
    }
    return distance
}

func GetBucketIndex(node [20]byte, secondNode [20]byte) int {
    distance := Distance(node, secondNode)

    var intDistance big.Int
    intDistance.SetBytes(distance[:])
    return intDistance.BitLen()
}

func (buckets *Buckets) AddNode(local [20]byte, remote [20]byte) ([20]byte, int, error) {
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

    return buckets.buckets[bucketIndex].Front().Value.([20]byte), bucketIndex, errors.New("Please ping this node")
}

func (buckets* Buckets) RemoveNode(local [20]byte, remote [20]byte) ([20]byte, int, error) {
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

func (buckets* Buckets) GetBucket(index int) *list.List {
    return &buckets.buckets[index]
}

func (buckets* Buckets) GetNodeInfo(nodeId [20]byte) (*NodeInfo, bool) {
    nodeInfo, found := buckets.nodes[nodeId]
    return nodeInfo, found
}
