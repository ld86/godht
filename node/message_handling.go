package node

import (
	"github.com/ld86/godht/messaging"
)

func (node *Node) DispatchMessage(message *messaging.Message) {
	node.addNodeToBuckets(message.FromId)

	switch message.Action {
	case "store_value":
		key := message.Ids[0]
		value := message.Payload
		node.storage.SetKey(key[:], value)
	case "ping":
		outputMessage := messaging.Message{FromId: node.id,
			ToId:          message.FromId,
			Action:        "pong",
			TransactionID: message.TransactionID,
		}
		node.messaging.SendMessage(outputMessage)
	case "pong":
		waitingTicket, found := node.waiting[message.FromId]
		if found {
			waitingTicket.GotPong = true
		}
	case "find_node":
		if len(message.Ids) == 0 {
			return
		}

		targetID := message.Ids[0]
		nearestIds := node.buckets.GetNearestIds(node.id, targetID, 5)
		outputMessage := messaging.Message{FromId: node.id,
			ToId:          message.FromId,
			Action:        "find_node_result",
			Ids:           nearestIds,
			TransactionID: message.TransactionID,
		}

		node.messaging.SendMessage(outputMessage)
	case "find_node_result":
		for _, nodeID := range message.Ids {
			node.addNodeToBuckets(nodeID)
		}
	}
}
