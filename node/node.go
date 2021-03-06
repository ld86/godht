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
	"github.com/ld86/godht/storage"
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
	storage   *storage.Storage

	defaultReceiver chan messaging.Message

	waiting map[types.NodeID]*WaitingTicket
}

func NewNodeWithId(id types.NodeID, bootstrap []string) *Node {
	return &Node{id: id,
		name:            petname.Generate(2, " "),
		bootstrap:       bootstrap,
		buckets:         buckets.NewBuckets(BucketSize),
		messaging:       messaging.NewMessaging(),
		storage:         storage.NewStorage(),
		defaultReceiver: make(chan messaging.Message, 100),
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

func (node *Node) checkNode(nodeID types.NodeID) {
	transaction := node.messaging.NewTransaction()
	defer transaction.Close()

	message := messaging.Message{FromId: node.id,
		ToId:   nodeID,
		Action: "ping",
		Ids:    []types.NodeID{node.id},
	}
	transaction.SendMessage(message)

	select {
	case _ = <-transaction.Receiver():
		log.Printf("Save %s in buckets\n", nodeID.String())
		node.addNodeToBuckets(nodeID)
	case <-time.After(3 * time.Second):
		log.Printf("Remove %s from buckets\n", nodeID.String())
		node.buckets.RemoveNode(node.id, nodeID)
	}

}

func (node *Node) pingNodes() {
	for {
		node.messaging.Mapping.Range(func(k, _ interface{}) bool {
			go node.checkNode(k.(types.NodeID))
			return true
		})

		time.Sleep(5 * time.Second)
	}
}

func (node *Node) discoverNodes() {
	for {
		for _, nodeID := range node.FindNode(node.id) {
			node.addNodeToBuckets(nodeID)
		}
		time.Sleep(120 * time.Second)
	}
}

func (node *Node) redistributeKeys() {
	for {
		time.Sleep(5 * time.Second)

		oldestKey, oldestValue, oldestTime, err := node.storage.OldestElement()
		if err != nil {
			continue
		}
		if time.Now().Sub(oldestTime) < 5*time.Second {
			continue
		}

		node.StoreValue(oldestKey, oldestValue)
	}
}

func (node *Node) Serve() {
	go node.messaging.Serve()
	go node.buckets.Serve()
	go node.storage.Serve()

	go node.doBootstrap()
	go node.discoverNodes()
	go node.pingNodes()
	go node.redistributeKeys()

	node.messaging.SetReceiver(node.defaultReceiver)

	for {
		select {
		case message := <-node.messaging.Receiver():
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
		node.messaging.SendMessage(pingMessage)

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
