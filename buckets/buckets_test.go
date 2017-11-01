package buckets_test

import (
	"testing"
	"github.com/ld86/godht/node"
	"github.com/ld86/godht/buckets"
)

func TestAddNode(t* testing.T) {
	a := node.NewNode()
	b := node.NewNode()

	buckets := buckets.NewBuckets(10)
	c, err := buckets.AddNode(a, b)

	if err != nil {
		t.Error("All buckets must be available")
	}

	if c.Id() != b.Id() {
		t.Error("On success we should get remote node")
	}

}
