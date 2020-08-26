package buckets

import (
	"sort"

	"github.com/ld86/godht/types"
)

type nodeWithDistance struct {
	ID       types.NodeID
	Distance [20]byte
}

type nodesWithDistances []nodeWithDistance

type NodesWithDistances struct {
	id    types.NodeID
	nodes nodesWithDistances
}

func NewNodesWithDistances(id types.NodeID) *NodesWithDistances {
	return &NodesWithDistances{id: id,
		nodes: make([]nodeWithDistance, 0),
	}
}

func (nodes *NodesWithDistances) AddNode(id types.NodeID) {
	nodeAndDistance := nodeWithDistance{ID: id, Distance: Distance(nodes.id, id)}
	nodes.nodes = append(nodes.nodes, nodeAndDistance)
}

func (nodes *NodesWithDistances) Sort() {
	sort.Sort(nodes.nodes)
}

func (nodes *NodesWithDistances) GetID(rank int) types.NodeID {
	return nodes.nodes[rank].ID
}

func (nodes *NodesWithDistances) GetDistance(rank int) [20]byte {
	return nodes.nodes[rank].Distance
}

func (nodes *NodesWithDistances) Len() int {
	return len(nodes.nodes)
}

func (nodes nodesWithDistances) Len() int {
	return len(nodes)
}

func (nodes nodesWithDistances) Swap(i, j int) {
	nodes[i], nodes[j] = nodes[j], nodes[i]
}

func (nodes nodesWithDistances) Less(i, j int) bool {
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
		nodesAndDistances := NewNodesWithDistances(remote)

		for nodeId := range buckets.nodes {
			if nodeId == remote {
				continue
			}
			nodesAndDistances.AddNode(nodeId)
		}

		nodesAndDistances.Sort()

		for i := 0; i < nodesAndDistances.Len() && len(result) < k; i++ {
			result = append(result, nodesAndDistances.GetID(i))
		}
	} else {
		bucket := buckets.buckets[bucketIndex-1]

		for it := bucket.Back(); it != nil && len(result) < k; it = it.Prev() {
			result = append(result, it.Value.(types.NodeID))
		}
	}

	return result
}
