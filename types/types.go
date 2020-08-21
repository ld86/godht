package types

import "encoding/hex"

type NodeID [20]byte

func (nodeID *NodeID) String() string {
	return hex.EncodeToString(nodeID[:])
}
