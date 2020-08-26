package node

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/ld86/godht/buckets"
	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

func (node *Node) FindNode(targetNodeID types.NodeID) []types.NodeID {
	alpha := 3
	k := 10

	nearestIds := node.buckets.GetNearestIds(node.id, targetNodeID, alpha)
	alreadyQueried := make(map[types.NodeID]bool)
	nodesAndDistances := make(buckets.NodesAndDistances, 0)

	foundNodes := make([]types.NodeID, 0)

	fmt.Println("From buckets")
	for _, nearestID := range nearestIds {
		fmt.Println(nearestID.String())
		alreadyQueried[nearestID] = false
		nodeAndDistance := buckets.NodeAndDistance{ID: nearestID, Distance: buckets.Distance(targetNodeID, nearestID)}
		nodesAndDistances = append(nodesAndDistances, nodeAndDistance)
	}

	receivedNodeID := make(chan messaging.Message)

	found := true
	for found {
		found = false
		sort.Sort(nodesAndDistances)
		queried := 0
		for i := 0; i < len(nodesAndDistances) && i < k; i++ {
			fmt.Printf("%s %v\n", nodesAndDistances[i].ID.String(), alreadyQueried[nodesAndDistances[i].ID])
			if alreadyQueried[nodesAndDistances[i].ID] {
				continue
			}
			alreadyQueried[nodesAndDistances[i].ID] = true

			queried++
			go func(j int) {
				transactionID := types.NewTransactionID()
				transactionReceiver := make(chan messaging.Message)

				message := messaging.Message{FromId: node.id,
					ToId:          nodesAndDistances[j].ID,
					Action:        "find_node",
					Ids:           []types.NodeID{targetNodeID},
					TransactionID: &transactionID,
				}

				node.messaging.AddTransactionReceiver(transactionID, transactionReceiver)
				defer node.messaging.RemoveTransactionReceiver(transactionID)

				fmt.Printf("Asking %s\n", nodesAndDistances[j].ID.String())
				node.messaging.SendMessage(message)

				select {
				case response := <-transactionReceiver:
					receivedNodeID <- response

				case <-time.After(3 * time.Second):
					log.Println("Timeout")
				}
			}(i)
		}

		t := true
		for i := 0; i < queried && t; i++ {
			select {
			case response := <-receivedNodeID:
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
						nodeAndDistance := buckets.NodeAndDistance{ID: nodeID, Distance: buckets.Distance(nodeID, targetNodeID)}
						nodesAndDistances = append(nodesAndDistances, nodeAndDistance)
					}
				}
			case <-time.After(3 * time.Second):
				t = false
			}
		}

	}

	sort.Sort(nodesAndDistances)

	for i := 0; i < len(nodesAndDistances) && i < alpha; i++ {
		foundNodes = append(foundNodes, nodesAndDistances[i].ID)
	}

	return foundNodes
}
