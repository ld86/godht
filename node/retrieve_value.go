package node

import (
	"fmt"
	"log"
	"time"

	"github.com/ld86/godht/utils"

	"github.com/ld86/godht/buckets"
	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

func (node *Node) RetrieveValue(key []byte) []byte {
	alpha := 3
	k := 10

	keyID := utils.HashBytesToNodeID(key)

	nearestIds := node.buckets.GetNearestIds(node.id, keyID, alpha)
	alreadyQueried := make(map[types.NodeID]bool)
	nodesAndDistances := buckets.NewNodesWithDistances(keyID)

	fmt.Println("From buckets")
	for _, nearestID := range nearestIds {
		fmt.Println(nearestID.String())
		alreadyQueried[nearestID] = false
		nodesAndDistances.AddNode(nearestID)
	}

	receivedNodeID := make(chan messaging.Message)

	found := true
	var value []byte

	for found && value == nil {
		found = false
		nodesAndDistances.Sort()
		queried := 0
		for i := 0; i < nodesAndDistances.Len() && i < k; i++ {
			candidateID := nodesAndDistances.GetID(i)

			fmt.Printf("%s %v\n", candidateID.String(), alreadyQueried[candidateID])

			if alreadyQueried[candidateID] {
				continue
			}

			alreadyQueried[candidateID] = true

			queried++

			go func(candidateID types.NodeID) {
				transaction := node.messaging.NewTransaction()
				defer transaction.Close()

				message := messaging.Message{FromId: node.id,
					ToId:   candidateID,
					Action: "retrieve_value",
					Ids:    []types.NodeID{keyID},
				}

				fmt.Printf("Asking %s\n", candidateID.String())
				transaction.SendMessage(message)

				select {
				case response := <-transaction.Receiver():
					receivedNodeID <- response

				case <-time.After(3 * time.Second):
					log.Println("Timeout")
				}
			}(candidateID)
		}

		t := true
		for i := 0; i < queried && t; i++ {
			select {
			case response := <-receivedNodeID:
				if response.Payload != nil {
					value = append([]byte{}, response.Payload...)
				} else {
					for _, nodeID := range response.Ids {
						if nodeID == node.id {
							continue
						}

						found = true
						node.addNodeToBuckets(nodeID)
						_, f := alreadyQueried[nodeID]
						if !f {
							fmt.Printf("Added %s\n", nodeID.String())

							alreadyQueried[nodeID] = false
							nodesAndDistances.AddNode(nodeID)
						}
					}
				}
			case <-time.After(3 * time.Second):
				t = false
			}
		}

	}

	return value
}
