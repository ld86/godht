package buckets

import (
	"sort"
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

type NodeAndDistance struct {
	Id [20]byte
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
			return true;
		}
	}
	return false;
}

func (buckets* Buckets) GetNearestIds(local [20]byte, remote [20] byte, k int) [][20]byte {
    buckets.mutex.Lock()
    defer buckets.mutex.Unlock()

    result := make([][20]byte, 0)
    bucketIndex := GetBucketIndex(local, remote)

	if bucketIndex == 0 || buckets.buckets[bucketIndex - 1].Len() < k {
		nodesAndDistances := make(NodesAndDistances, 0)
		for nodeId := range buckets.nodes {
			nodeAndDistance := NodeAndDistance{Id: nodeId, Distance: Distance(nodeId, remote)}
			nodesAndDistances = append(nodesAndDistances, nodeAndDistance)
		}

		sort.Sort(nodesAndDistances)

		for i := 0; i < len(nodesAndDistances) && len(result) < k; i++ {
			result = append(result, nodesAndDistances[i].Id)
		}
	} else {
		bucket := buckets.buckets[bucketIndex - 1]

        for it := bucket.Back(); it != nil && len(result) < k; it = it.Prev() {
            result = append(result, it.Value.([20]byte))
        }
    }

    return result
}

