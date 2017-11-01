package node_test

import (
    "testing"
    "math/big"
    "github.com/ld86/godht/node"
)

func BytesToBigInt(bytes [20]byte) *big.Int {
    result := big.NewInt(0)
    result.SetBytes(bytes[:])
    return result
}

func TestDistance(t *testing.T) {
    a := node.NewNode()
    b := node.NewNode()
    c := node.NewNode()
    zero := big.NewInt(0)

    if BytesToBigInt(a.Distance(b)).Cmp(BytesToBigInt(b.Distance(a))) != 0 {
        t.Error("Distance must be symmetric")
    }

    if BytesToBigInt(a.Distance(a)).Cmp(zero) != 0 {
        t.Error("Distance to youself must be zero")
    }

    if BytesToBigInt(a.Distance(b)).Cmp(zero) == 0 {
        t.Error("This should never happens")
    }

    ab := BytesToBigInt(a.Distance(b))
    bc := BytesToBigInt(b.Distance(c))
    ac := BytesToBigInt(a.Distance(c))

    if big.NewInt(0).Add(ab, bc).Cmp(ac) == -1 {
        t.Error("Triangle property failed")
    }
}
