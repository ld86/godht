package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/ld86/godht/node"
	"github.com/ld86/godht/types"

	prompt "github.com/c-bata/go-prompt"
)

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
			nodeId := it.Value.(types.NodeID)
			nodeInfoFromBuckets, found := currentNode.Buckets().GetNodeInfo(nodeId)
			if !found {
				return
			}
			fmt.Printf("%s %s\n", nodeId.String(), nodeInfoFromBuckets.UpdateTime)
		}
	}
	return
}

func completer(t prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "print"},
		{Text: "new"},
		{Text: "serve"},
	}
}

func main() {
	log.SetOutput(ioutil.Discard)
	p := prompt.New(
		executor,
		completer,
	)
	p.Run()
}
