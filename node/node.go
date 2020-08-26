package node

import (
	"crypto/rand"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	petname "github.com/dustinkirkland/golang-petname"

	"github.com/ld86/godht/buckets"
	"github.com/ld86/godht/messaging"
	"github.com/ld86/godht/types"
)

const BucketSize = 10

type WaitingTicket struct {
	GotPong bool
}

type Node struct {
	id        types.NodeID
	name      string
	bootstrap []string

	buckets   *buckets.Buckets
	messaging *messaging.Messaging

	defaultReceiver chan messaging.Message

	waiting map[types.NodeID]*WaitingTicket
}

func NewNodeWithId(id types.NodeID, bootstrap []string) *Node {
	return &Node{id: id,
		name:            petname.Generate(2, " "),
		bootstrap:       bootstrap,
		buckets:         buckets.NewBuckets(BucketSize),
		messaging:       messaging.NewMessaging(),
		defaultReceiver: make(chan messaging.Message),
		waiting:         make(map[types.NodeID]*WaitingTicket)}
}

func NewNode(bootstrap []string) *Node {
	var id types.NodeID
	_, err := rand.Read(id[:])
	if err != nil {
		log.Panicf("rand.Read failed, %s", err)
	}
	return NewNodeWithId(id, bootstrap)
}

func (node *Node) String() string {
	return fmt.Sprintf("%s %s %s",
		node.Name(),
		node.messaging.GetLocalAddr(),
		node.id.String())
}

func (node *Node) Id() types.NodeID {
	return node.id
}

func (node *Node) Name() string {
	return node.name
}

func (node *Node) Buckets() *buckets.Buckets {
	return node.buckets
}

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
		node.messaging.MessagesToSend <- message

		select {
		case response := <-transactionReceiver:
			node.addNodeToBuckets(response.FromId)
			for _, nodeID := range response.Ids {
				if nodeID == node.id {
					continue
				}
				node.addNodeToBuckets(nodeID)
			}

			for _, nodeID2 := range node.FindNode(node.id) {
				node.addNodeToBuckets(nodeID2)
			}

		case <-time.After(3 * time.Second):
			log.Println("Timeout")
		}

	}
}

func (node *Node) pingOldNodes() {
	for {
		for i := 0; i < 160; i++ {
			bucket := node.buckets.GetBucket(i)
			if bucket.Len() > 0 {
				message := messaging.Message{FromId: node.id,
					ToId:   bucket.Front().Value.(types.NodeID),
					Action: "find_node",
					Ids:    []types.NodeID{node.id},
				}
				node.messaging.MessagesToSend <- message

			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) Serve() {
	go node.messaging.Serve()
	go node.buckets.Serve()

	go node.doBootstrap()
	go node.pingOldNodes()

	node.messaging.SetDefaultReceiver(node.defaultReceiver)

	for {
		select {
		case message := <-node.defaultReceiver:
			node.DispatchMessage(&message)
		}
	}
}

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
		nodeAndDistance := buckets.NodeAndDistance{Id: nearestID, Distance: buckets.Distance(targetNodeID, nearestID)}
		nodesAndDistances = append(nodesAndDistances, nodeAndDistance)
	}

	receivedNodeID := make(chan messaging.Message)

	found := true
	for found {
		found = false
		sort.Sort(nodesAndDistances)
		queried := 0
		for i := 0; i < len(nodesAndDistances) && i < k; i++ {
			fmt.Printf("%s %v\n", nodesAndDistances[i].Id.String(), alreadyQueried[nodesAndDistances[i].Id])
			if alreadyQueried[nodesAndDistances[i].Id] {
				continue
			}
			alreadyQueried[nodesAndDistances[i].Id] = true

			queried++
			go func(j int) {
				transactionID := types.NewTransactionID()
				transactionReceiver := make(chan messaging.Message)

				message := messaging.Message{FromId: node.id,
					ToId:          nodesAndDistances[j].Id,
					Action:        "find_node",
					Ids:           []types.NodeID{targetNodeID},
					TransactionID: &transactionID,
				}

				node.messaging.AddTransactionReceiver(transactionID, transactionReceiver)
				defer node.messaging.RemoveTransactionReceiver(transactionID)

				fmt.Printf("Asking %s\n", nodesAndDistances[j].Id.String())
				node.messaging.MessagesToSend <- message

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
						nodeAndDistance := buckets.NodeAndDistance{Id: nodeID, Distance: buckets.Distance(nodeID, targetNodeID)}
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
		foundNodes = append(foundNodes, nodesAndDistances[i].Id)
	}

	return foundNodes
}

func (node *Node) addNodeToBuckets(fromId types.NodeID) {
	returnedNodeId, bucketIndex, err := node.buckets.AddNode(node.id, fromId)

	if err == nil {
		log.Printf("Successfuly add remote %s to buckets", fromId.String())
		return
	}

	if err != nil && bucketIndex == -1 {
		return
	}

	log.Printf("Bucket %d is full, trying to ping node %s", bucketIndex, returnedNodeId.String())
	go func() {
		waitingTicket := &WaitingTicket{GotPong: false}

		{
			mutex := sync.Mutex{}
			mutex.Lock()
			defer mutex.Unlock()
			if _, found := node.waiting[returnedNodeId]; found {
				log.Printf("Already waiting for node %s", returnedNodeId.String())
				return
			}
			node.waiting[returnedNodeId] = waitingTicket
		}

		pingMessage := messaging.Message{FromId: node.id, ToId: returnedNodeId, Action: "ping"}
		node.messaging.MessagesToSend <- pingMessage

		log.Printf("Waiting 5 seconds for %v", returnedNodeId)
		time.Sleep(5 * time.Second)

		{
			mutex := sync.Mutex{}
			mutex.Lock()
			defer mutex.Unlock()

			if !waitingTicket.GotPong {
				log.Printf("Did not get pong from %s, removing it", returnedNodeId.String())
				node.buckets.RemoveNode(node.id, returnedNodeId)
				_, _, err = node.buckets.AddNode(node.id, fromId)
				if err != nil {
					log.Fatalf("Something goes wrong with buckets")
				}
			} else {
				log.Printf("Got pong from %s, leave it in buckets", returnedNodeId.String())
			}
			delete(node.waiting, returnedNodeId)
		}
	}()
}

func (node *Node) DispatchMessage(message *messaging.Message) {
	log.Printf("Got message %s", message.String())

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

		targetId := message.Ids[0]
		nearestIds := node.buckets.GetNearestIds(node.id, targetId, 5)
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
