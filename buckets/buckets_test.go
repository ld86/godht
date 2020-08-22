package buckets_test

import (
	"math"
	"math/big"
	"sort"
	"testing"

	"github.com/ld86/godht/buckets"
	"github.com/ld86/godht/node"
	"github.com/ld86/godht/types"
)

func TestGetBucketIndex(t *testing.T) {
	a := node.NewNode([]string{})
	b := node.NewNode([]string{})

	if buckets.GetBucketIndex(a.Id(), a.Id()) != 0 {
		t.Error("Two same nodes should be placed in one bucket")
	}

	if buckets.GetBucketIndex(a.Id(), b.Id()) > 160 {
		t.Error("Bucket index must be less than 160")
	}

	for j := 0; j < 20; j++ {
		for i := 1; i < 2; i++ {
			var manualId [20]byte

			c := node.NewNodeWithId(manualId, []string{})
			manualId[19-j] = byte(i)
			d := node.NewNodeWithId(manualId, []string{})

			if buckets.GetBucketIndex(c.Id(), d.Id()) != int(math.Log2(float64(i))+1)+j*8 {
				t.Error("Bad bucket index")
			}
		}
	}
}

func TestDistance(t *testing.T) {
	var nodesAndDistances buckets.NodesAndDistances
	for j := 0; j < 20; j++ {
		for i := 1; i < 2; i++ {
			var manualId [20]byte

			a := node.NewNodeWithId(manualId, []string{})
			manualId[j] = byte(i)
			b := node.NewNodeWithId(manualId, []string{})

			distance := buckets.Distance(a.Id(), b.Id())
			nodesAndDistances = append(nodesAndDistances, buckets.NodeAndDistance{Id: b.Id(), Distance: distance})
		}
	}
	sort.Sort(nodesAndDistances)
	for j := 0; j < 19; j++ {
		var a, b big.Int
		a.SetBytes(nodesAndDistances[j].Distance[:])
		b.SetBytes(nodesAndDistances[j+1].Distance[:])
		if a.Cmp(&b) > 0 {
			t.Errorf("Wrong order %s > %s", a.String(), b.String())
		}
	}
}

func TestAddNode(t *testing.T) {
	const bucketSize = 10

	local := node.NewNode([]string{})
	buckets := buckets.NewBuckets(bucketSize)

	go buckets.Serve()

	remote := node.NewNode([]string{})
	_, bucketIndex, _ := buckets.AddNode(local.Id(), remote.Id())

	var overflowedError error
	var lastBucket int
	var lastReturned [20]byte

	for overflowedError == nil {
		remote := node.NewNode([]string{})
		lastReturned, lastBucket, overflowedError = buckets.AddNode(local.Id(), remote.Id())

		if lastBucket == -1 {
			t.Error("This should never happens")
		}

		if overflowedError == nil && lastReturned != remote.Id() {
			t.Error("On success we should get remote node")
		}
	}

	bucket := buckets.GetBucket(lastBucket)
	if bucket.Len() != bucketSize {
		t.Error("This bucket should overflowed")
	}

	if bucket.Front().Value.(types.NodeID) != lastReturned {
		t.Error("On overflow we should ping first node from bucket")
	}

	bucket = buckets.GetBucket(bucketIndex)
	if bucket.Len() == 0 {
		t.Error("Bucket must be not empty")
	}

	if bucket.Front().Value.(types.NodeID) != remote.Id() {
		t.Error("Remote node should be first")
	}

	_, _, err := buckets.AddNode(local.Id(), remote.Id())
	if err != nil {
		t.Error("We already have remote in buckets, so we need to move it tail")
	}

	if bucket.Back().Value.(types.NodeID) != remote.Id() {
		t.Error("Remote node should be last")
	}

	oldSize := bucket.Len()
	removedNode, _, err := buckets.RemoveNode(local.Id(), remote.Id())
	if err != nil {
		t.Error("Should never happens")
	}

	if removedNode != remote.Id() {
		t.Error("Removed node should be returned")
	}

	if oldSize-1 != bucket.Len() {
		t.Error("Bucket should contain less nodes")
	}

}
