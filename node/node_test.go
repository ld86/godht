package node_test

import (
    "math/big"
    "testing"
    "github.com/ld86/godht/node"
    "github.com/ld86/godht/buckets"
)

func BytesToBigInt(bytes [20]byte) *big.Int {
    result := big.NewInt(0)
    result.SetBytes(bytes[:])
    return result
}

func TestDistance(t *testing.T) {
    a := node.NewNode([]string{})
    b := node.NewNode([]string{})
    c := node.NewNode([]string{})
    zero := big.NewInt(0)

    if BytesToBigInt(buckets.Distance(a.Id(), b.Id())).Cmp(BytesToBigInt(buckets.Distance(b.Id(), a.Id()))) != 0 {
        t.Error("Distance must be symmetric")
    }

    if BytesToBigInt(buckets.Distance(a.Id(), a.Id())).Cmp(zero) != 0 {
        t.Error("Distance to youself must be zero")
    }

    if BytesToBigInt(buckets.Distance(a.Id(), b.Id())).Cmp(zero) == 0 {
        t.Error("This should never happens")
    }

    ab := BytesToBigInt(buckets.Distance(a.Id(), b.Id()))
    bc := BytesToBigInt(buckets.Distance(b.Id(), c.Id()))
    ac := BytesToBigInt(buckets.Distance(a.Id(), c.Id()))

    if big.NewInt(0).Add(ab, bc).Cmp(ac) == -1 {
        t.Error("Triangle property failed")
    }
}
