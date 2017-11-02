package node

import (
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

func (buckets *Buckets) AddNode(local *Node, remote *Node) (*Node, int, error) {
    bucketIndex := local.GetBucketIndex(remote)

    if bucketIndex == 0 {
        return local, -1, errors.New("Cannot add yourself to buckets")
    }

    bucketIndex--

    remoteElement, ok := buckets.nodes[bucketIndex][remote.Id()]

    if !ok && buckets.buckets[bucketIndex].Len() < buckets.k {
        e := buckets.buckets[bucketIndex].PushBack(remote)
        buckets.nodes[bucketIndex][remote.Id()] = e
        return remote, bucketIndex, nil
    }

    if ok {
        buckets.buckets[bucketIndex].MoveToBack(remoteElement)
        return remote, bucketIndex, nil
    }

    return buckets.buckets[bucketIndex].Front().Value.(*Node), bucketIndex, errors.New("Please ping this node")
}

func (buckets* Buckets) RemoveNode(local *Node, remote *Node) (*Node, int, error) {
    bucketIndex := local.GetBucketIndex(remote)

    if bucketIndex == 0 {
        return local, -1, errors.New("Cannot remove yourself from buckets")
    }

    bucketIndex--
    remoteElement, ok := buckets.nodes[bucketIndex][remote.Id()]

    if !ok {
        return remote, bucketIndex, nil
    }

    delete(buckets.nodes[bucketIndex], remote.Id())
    buckets.buckets[bucketIndex].Remove(remoteElement)

    return remote, bucketIndex, nil
}

func (buckets* Buckets) GetBucket(index int) *list.List {
    return &buckets.buckets[index]
}


