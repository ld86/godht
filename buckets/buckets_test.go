package buckets_test

import (
	"testing"
	"github.com/ld86/godht/node"
	"github.com/ld86/godht/buckets"
)

func TestAddNode(t* testing.T) {
	local := node.NewNode()
	buckets := buckets.NewBuckets(1000)

	remote := node.NewNode()
	_, bucketIndex, _ := buckets.AddNode(local, remote)

	for i := 0; i < 500; i++ {
		remote := node.NewNode()
		remoteReturned, _, err := buckets.AddNode(local, remote)

		if err != nil {
			t.Error("All buckets must be available")
		}

		if remoteReturned.Id() != remote.Id() {
			t.Error("On success we should get remote node")
		}
	}

	bucket := buckets.GetBucket(bucketIndex)
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
}
