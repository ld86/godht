package node

import (
	"errors"

	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"

	"github.com/ld86/godht/utils"
)

func (node *Node) StoreValue(key []byte, value []byte) error {
	keyID := utils.HashBytesToNodeID(key)

	nearestNodes := node.FindNode(keyID)

	if len(nearestNodes) == 0 {
		return errors.New("Cannot find any nearest nodes")
	}

	for _, nodeID := range nearestNodes {
		message := messaging.Message{FromId: node.id,
			ToId:    nodeID,
			Action:  "store_value",
			Ids:     []types.NodeID{keyID},
			Payload: value,
		}

		node.messaging.SendMessage(message)
	}

	return nil
}
