package node_test

import (
    "testing"
    "math"
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

func TestGetBucketIndex(t* testing.T) {
    a := node.NewNode()
    b := node.NewNode()

    if a.GetBucketIndex(a) != 0 {
        t.Error("Two same nodes should be placed in one bucket")
    }

    if a.GetBucketIndex(b) > 160 {
        t.Error("Bucket index must be less than 160")
    }

    for j := 0; j < 20; j++ {
        for i := 1; i < 2; i++ {
            var manualId [20]byte

            c := node.NewNodeWithId(manualId)
            manualId[19 - j] = byte(i)
            d := node.NewNodeWithId(manualId)

            if c.GetBucketIndex(d) != int(math.Log2(float64(i)) + 1) + j * 8 {
                t.Error("Bad bucket index")
            }
        }
    }
}
