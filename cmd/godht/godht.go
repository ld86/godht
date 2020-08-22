package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	math_rand "math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/ld86/godht/node"
	"github.com/ld86/godht/types"

	prompt "github.com/c-bata/go-prompt"
)

func init() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("init()")
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(ioutil.Discard)
}

var currentNode *node.Node

func executor(line string) {
	line = strings.TrimSpace(line)
	fields := strings.Split(line, " ")

	if len(fields) == 0 {
		return
	}

	switch fields[0] {
	case "print":
		fmt.Println(currentNode)
		break
	case "new":
		var bootstrapServers []string
		if len(fields) > 1 {
			bootstrapServers = append(bootstrapServers, fields[1:]...)
		}

		currentNode = node.NewNode(bootstrapServers)

		fmt.Println(currentNode)
	case "serve":
		go currentNode.Serve()
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
	case "buckets":
		for k, v := range currentNode.Buckets().GetSizes() {
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
		bucket := currentNode.Buckets().GetBucket(bucketIndex)
		if bucket.Len() == 0 {
			return
		}
		for it := bucket.Front(); it != nil; it = it.Next() {
			nodeID := it.Value.(types.NodeID)
			nodeInfoFromBuckets, found := currentNode.Buckets().GetNodeInfo(nodeID)
			if !found {
				return
			}
			fmt.Printf("%s %s\n", nodeID.String(), nodeInfoFromBuckets.UpdateTime)
		}
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "print"},
		{Text: "new"},
		{Text: "serve"},
		{Text: "logs"},
		{Text: "buckets"},
		{Text: "bucket"},
	}
}

func main() {
	p := prompt.New(
		executor,
		completer,
	)
	p.Run()
}
