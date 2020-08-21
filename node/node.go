package node

import (
	"crypto/rand"
	"fmt"
	"log"
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

	waiting map[[20]byte]*WaitingTicket
}

func NewNodeWithId(id types.NodeID, bootstrap []string) *Node {
	return &Node{id: id,
		name:      petname.Generate(2, " "),
		bootstrap: bootstrap,
		buckets:   buckets.NewBuckets(BucketSize),
		messaging: messaging.NewMessaging(),
		waiting:   make(map[[20]byte]*WaitingTicket)}
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
	for _, remoteIP := range node.bootstrap {
		message := messaging.Message{FromId: node.id,
			Action: "find_node",
			IpAddr: &remoteIP,
			Ids:    []types.NodeID{node.id},
		}
		node.messaging.OutputMessages <- message
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
				node.messaging.OutputMessages <- message
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) Serve() {
	go node.messaging.Serve()
	go node.doBootstrap()
	go node.pingOldNodes()
	for {
		select {
		case message := <-node.messaging.InputMessages:
			node.DispatchMessage(&message)
		}
	}
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
		node.messaging.OutputMessages <- pingMessage

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
		node.messaging.OutputMessages <- outputMessage
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
		nearestIds := node.buckets.GetNearestIds(node.id, targetId, 3)
		outputMessage := messaging.Message{FromId: node.id,
			ToId:   message.FromId,
			Action: "find_node_result",
			Ids:    nearestIds}

		node.messaging.OutputMessages <- outputMessage
	case "find_node_result":
		for _, nodeId := range message.Ids {
			node.addNodeToBuckets(nodeId)
		}
	}

}
