package node

import (
	"errors"

	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

func (node *Node) StoreValue(key types.NodeID, value []byte) error {
	err := node.storage.SetKey(key, value)

	if err != nil {
		return err
	}

	nearestNodes := node.FindNode(key)

	if len(nearestNodes) == 0 {
		return errors.New("Cannot find any nearest nodes")
	}

	for _, nodeID := range nearestNodes {
		message := messaging.Message{FromId: node.id,
			ToId:    nodeID,
			Action:  "store_value",
			Ids:     []types.NodeID{key},
			Payload: value,
		}

		node.messaging.SendMessage(message)
	}

	return nil
}
