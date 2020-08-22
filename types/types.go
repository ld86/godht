package types

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

type NodeID [20]byte

func (nodeID *NodeID) String() string {
	return hex.EncodeToString(nodeID[:])
}

type TransactionID [20]byte

func NewTransactionID() TransactionID {
	var id TransactionID
	_, err := rand.Read(id[:])
	if err != nil {
		log.Panicf("rand.Read failed, %s", err)
	}
	return id
}

func (transactionID *TransactionID) String() string {
	return hex.EncodeToString(transactionID[:])
}
