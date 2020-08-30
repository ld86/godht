package node

import (
	"log"
	"time"

	"github.com/ld86/godht/buckets"
	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

func (node *Node) FindNode(targetNodeID types.NodeID) []types.NodeID {
	alpha := 3
	k := 10

	nearestIds := node.buckets.GetNearestIds(node.id, targetNodeID, alpha)
	nodesAndDistances := buckets.NewNodesWithDistances(targetNodeID)

	alreadyQueried := make(map[types.NodeID]bool)
	gotResponse := make(map[types.NodeID]bool)

	foundNodes := make([]types.NodeID, 0)

	log.Println("From buckets")
	for _, nearestID := range nearestIds {
		log.Println(nearestID.String())
		alreadyQueried[nearestID] = false
		nodesAndDistances.AddNode(nearestID)
	}

	receivedNodeID := make(chan messaging.Message)

	found := true
	for found {
		found = false
		nodesAndDistances.Sort()
		queried := 0
		for i := 0; i < nodesAndDistances.Len() && i < k; i++ {
			candidateID := nodesAndDistances.GetID(i)
			log.Printf("%s %v %v\n", candidateID.String(), alreadyQueried[candidateID], gotResponse[candidateID])

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
					Action: "find_node",
					Ids:    []types.NodeID{targetNodeID},
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
				found = true
				gotResponse[response.FromId] = true

				for _, nodeID := range response.Ids {
					if nodeID == node.id {
						continue
					}

					node.addNodeToBuckets(nodeID)
					_, f := alreadyQueried[nodeID]
					if !f {
						log.Printf("Added %s\n", nodeID.String())

						alreadyQueried[nodeID] = false
						nodesAndDistances.AddNode(nodeID)

					}
				}
			case <-time.After(3 * time.Second):
				t = false
			}
		}

	}

	nodesAndDistances.Sort()

	for i := 0; i < nodesAndDistances.Len() && len(foundNodes) < alpha; i++ {
		ID := nodesAndDistances.GetID(i)
		if gotResponse[ID] {
			foundNodes = append(foundNodes, ID)
		}
	}

	return foundNodes
}
