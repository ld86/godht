package buckets

import (
	"sort"

	"github.com/ld86/godht/types"
)

type NodeAndDistance struct {
	ID       types.NodeID
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
		if nodes[i].Distance[index] == nodes[j].Distance[index] {
			continue
		}
		return nodes[i].Distance[index] < nodes[j].Distance[index]
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

			nodeAndDistance := NodeAndDistance{ID: nodeId, Distance: Distance(nodeId, remote)}
			nodesAndDistances = append(nodesAndDistances, nodeAndDistance)
		}

		sort.Sort(nodesAndDistances)

		for i := 0; i < len(nodesAndDistances) && len(result) < k; i++ {
			result = append(result, nodesAndDistances[i].ID)
		}
	} else {
		bucket := buckets.buckets[bucketIndex-1]

		for it := bucket.Back(); it != nil && len(result) < k; it = it.Prev() {
			result = append(result, it.Value.(types.NodeID))
		}
	}

	return result
}
