package node

import (
	"log"
	"time"

	"github.com/ld86/godht/buckets"
	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

func (node *Node) RetrieveValue(key types.NodeID) []byte {
	localValue, err := node.storage.GetKey(key)
	if err == nil {
		return localValue
	}

	alpha := 3
	k := 10

	nearestIds := node.buckets.GetNearestIds(node.id, key, alpha)
	alreadyQueried := make(map[types.NodeID]bool)
	nodesAndDistances := buckets.NewNodesWithDistances(key)

	log.Println("From buckets")
	for _, nearestID := range nearestIds {
		log.Println(nearestID.String())
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

			log.Printf("%s %v\n", candidateID.String(), alreadyQueried[candidateID])

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
					Ids:    []types.NodeID{key},
				}

				log.Printf("Asking %s\n", candidateID.String())
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
					node.addNodeToBuckets(response.FromId)
					for _, nodeID := range response.Ids {
						if nodeID == node.id {
							continue
						}

						found = true
						_, f := alreadyQueried[nodeID]
						if !f {
							log.Printf("Added %s\n", nodeID.String())

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
