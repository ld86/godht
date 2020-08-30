package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	math_rand "math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/ld86/godht/node"
	"github.com/ld86/godht/types"
	"github.com/ld86/godht/utils"

	prompt "github.com/c-bata/go-prompt"
)

func init() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("init()")
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

var mainNode *node.Node

func executor(line string) {
	line = strings.TrimSpace(line)
	fields := strings.Split(line, " ")

	if len(fields) == 0 {
		return
	}

	switch fields[0] {
	case "print":
		fmt.Println(mainNode)
		break
	case "new":
		var bootstrapServers []string
		if len(fields) > 1 {
			bootstrapServers = append(bootstrapServers, fields[1:]...)
		}

		mainNode = node.NewNode(bootstrapServers)

		fmt.Println(mainNode)
	case "serve":
		go mainNode.Serve()
	case "logs":
		i, err := strconv.Atoi(fields[1])
		if err != nil {
			return
		}
		if i == 0 {
			log.SetOutput(ioutil.Discard)
		} else {
			log.SetOutput(os.Stderr)
		}
	case "find":
		if len(fields) < 2 {
			return
		}
		keyID := utils.HashStringToNodeID(fields[1])
		nearestNodes := mainNode.FindNode(keyID)

		fmt.Println("Nearest nodes")
		for _, nodeID := range nearestNodes {
			fmt.Println(nodeID.String())
		}
	case "store":
		if len(fields) < 3 {
			return
		}

		key := []byte(fields[1])
		value := []byte(fields[2])

		err := mainNode.StoreValue(key, value)

		if err != nil {
			fmt.Println(err)
		}

	case "retrieve":
		if len(fields) < 2 {
			return
		}

		key := []byte(fields[1])

		value := mainNode.RetrieveValue(key)

		if value != nil {
			fmt.Println(string(value))
		}

	case "buckets":
		for k, v := range mainNode.Buckets().GetSizes() {
			fmt.Printf("%d %d\n", k, v)
		}
	case "bucket":
		if len(fields) < 2 {
			return
		}
		bucketIndex, err := strconv.Atoi(fields[1])
		if err != nil {
			return
		}
		bucket := mainNode.Buckets().GetBucket(bucketIndex)
		if bucket.Len() == 0 {
			return
		}
		for it := bucket.Front(); it != nil; it = it.Next() {
			nodeID := it.Value.(types.NodeID)
			nodeInfoFromBuckets, found := mainNode.Buckets().GetNodeInfo(nodeID)
			if !found {
				return
			}
			fmt.Printf("%s %s\n", nodeID.String(), nodeInfoFromBuckets.UpdateTime)
		}
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func main() {
	interactive := flag.Bool("i", false, "Start interactive shell")
	flag.Parse()

	if *interactive {
		p := prompt.New(
			executor,
			completer,
		)
		p.Run()
	} else {
		var bootstrapServers []string
		if len(flag.Args()) > 0 {
			bootstrapServers = append(bootstrapServers, flag.Args()...)
		}
		mainNode = node.NewNode(bootstrapServers)
		log.Println(mainNode)
		mainNode.Serve()
	}

}
