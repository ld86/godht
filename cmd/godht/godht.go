package main

import (
	"os"
    "github.com/ld86/godht/node"
)


func main() {
	var bootstrapServers []string
	if len(os.Args) > 1{
		bootstrapServers = append(bootstrapServers, os.Args[1:]...)
	}

    me := node.NewNode(bootstrapServers)
	me.Serve()
}
