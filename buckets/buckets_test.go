package buckets_test

import (
	"testing"
	"github.com/ld86/godht/node"
	"github.com/ld86/godht/buckets"
)

func TestAddNode(t* testing.T) {
	const bucketSize = 10

	local := node.NewNode()
	buckets := buckets.NewBuckets(bucketSize)

	remote := node.NewNode()
	_, bucketIndex, _ := buckets.AddNode(local, remote)

	var overflowedError error
	var lastBucket int
	var lastReturned *node.Node

	for overflowedError == nil {
		remote := node.NewNode()
		lastReturned, lastBucket, overflowedError = buckets.AddNode(local, remote)

		if lastBucket == -1 {
			t.Error("This should never happens")
		}

		if overflowedError == nil && lastReturned.Id() != remote.Id() {
			t.Error("On success we should get remote node")
		}
	}

	bucket := buckets.GetBucket(lastBucket)
	if bucket.Len() != bucketSize{
		t.Error("This bucket should overflowed")
	}

	if bucket.Front().Value.(*node.Node).Id() != lastReturned.Id() {
		t.Error("On overflow we should ping first node from bucket")
	}

	bucket = buckets.GetBucket(bucketIndex)
	if bucket.Len() == 0 {
		t.Error("Bucket must be not empty")
	}

	if bucket.Front().Value.(*node.Node).Id() != remote.Id() {
		t.Error("Remote node should be first")
	}

	_, _, err := buckets.AddNode(local, remote)
	if err != nil {
		t.Error("We already have remote in buckets, so we need to move it tail")
	}

	if bucket.Back().Value.(*node.Node).Id() != remote.Id() {
		t.Error("Remote node should be last")
	}

	oldSize := bucket.Len()
	removedNode, _, err := buckets.RemoveNode(local, remote)
	if err != nil {
		t.Error("Should never happens")
	}

	if removedNode.Id() != remote.Id() {
		t.Error("Removed node should be returned")
	}

	if oldSize - 1 != bucket.Len() {
		t.Error("Bucket should contain less nodes")
	}

}
