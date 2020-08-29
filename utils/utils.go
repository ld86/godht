package utils

import (
	"crypto/sha1"

	"github.com/ld86/godht/types"
)

func HashStringToNodeID(s string) types.NodeID {
	return HashBytesToNodeID([]byte(s))
}

func HashBytesToNodeID(b []byte) types.NodeID {
	hash := HashBytes(b)
	nodeID := types.NodeID{}
	copy(nodeID[:], hash[:])
	return nodeID
}

func HashString(s string) []byte {
	return HashBytes([]byte(s))
}

func HashBytes(b []byte) []byte {
	hasher := sha1.New()

	hasher.Write(b)
	return hasher.Sum(nil)
}
