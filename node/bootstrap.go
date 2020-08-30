package node

import (
	"log"
	"time"

	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

func (node *Node) doBootstrap() {
	if len(node.bootstrap) == 0 {
		return
	}

	transactionID := types.NewTransactionID()
	transactionReceiver := make(chan messaging.Message)
	node.messaging.AddTransactionReceiver(transactionID, transactionReceiver)
	defer node.messaging.RemoveTransactionReceiver(transactionID)

	for _, remoteIP := range node.bootstrap {
		message := messaging.Message{FromId: node.id,
			Action:        "find_node",
			IpAddr:        &remoteIP,
			Ids:           []types.NodeID{node.id},
			TransactionID: &transactionID,
		}
		node.messaging.SendMessage(message)

		select {
		case response := <-transactionReceiver:
			node.addNodeToBuckets(response.FromId)
			for _, nodeID := range node.FindNode(node.id) {
				node.addNodeToBuckets(nodeID)
			}

		case <-time.After(3 * time.Second):
			log.Println("Timeout")
		}

	}
}
