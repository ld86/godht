package node

import (
	"fmt"

	"github.com/ld86/godht/messaging"
)

func (node *Node) DispatchMessage(message *messaging.Message) {
	node.addNodeToBuckets(message.FromId)

	switch message.Action {
	case "ping":
		outputMessage := messaging.Message{FromId: node.id, ToId: message.FromId, Action: "pong"}
		node.messaging.MessagesToSend <- outputMessage
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
		fmt.Println(message.TransactionID)
		outputMessage := messaging.Message{FromId: node.id,
			ToId:          message.FromId,
			Action:        "find_node_result",
			Ids:           nearestIds,
			TransactionID: message.TransactionID,
		}

		node.messaging.MessagesToSend <- outputMessage
	case "find_node_result":
		for _, nodeId := range message.Ids {
			node.addNodeToBuckets(nodeId)
		}
	}
}
