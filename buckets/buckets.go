package buckets

import (
    "math/big"
    "errors"
    "container/list"
)

type Buckets struct {
    k int
    buckets [160]list.List
    nodes [160]map[[20]byte]*list.Element
}

func NewBuckets(k int) *Buckets {
    var nodes [160]map[[20]byte]*list.Element
    for i := 0; i < 160; i++ {
        nodes[i] = make(map[[20]byte]*list.Element, k)
    }
    return &Buckets{k: k, nodes: nodes}
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
    bucketIndex := GetBucketIndex(local, remote)

    if bucketIndex == 0 {
        return local, -1, errors.New("Cannot add yourself to buckets")
    }

    bucketIndex--

    remoteElement, ok := buckets.nodes[bucketIndex][remote]

    if !ok && buckets.buckets[bucketIndex].Len() < buckets.k {
        e := buckets.buckets[bucketIndex].PushBack(remote)
        buckets.nodes[bucketIndex][remote] = e
        return remote, bucketIndex, nil
    }

    if ok {
        buckets.buckets[bucketIndex].MoveToBack(remoteElement)
        return remote, bucketIndex, nil
    }

    return buckets.buckets[bucketIndex].Front().Value.([20]byte), bucketIndex, errors.New("Please ping this node")
}

func (buckets* Buckets) RemoveNode(local [20]byte, remote [20]byte) ([20]byte, int, error) {
    bucketIndex := GetBucketIndex(local, remote)

    if bucketIndex == 0 {
        return local, -1, errors.New("Cannot remove yourself from buckets")
    }

    bucketIndex--
    remoteElement, ok := buckets.nodes[bucketIndex][remote]

    if !ok {
        return remote, bucketIndex, nil
    }

    delete(buckets.nodes[bucketIndex], remote)
    buckets.buckets[bucketIndex].Remove(remoteElement)

    return remote, bucketIndex, nil
}

func (buckets* Buckets) GetBucket(index int) *list.List {
    return &buckets.buckets[index]
}
