package main

import (
	"os"

	"github.com/ld86/godht/node"
	"github.com/ld86/godht/viewer"
)

func main() {
	var bootstrapServers []string
	if len(os.Args) > 1 {
		bootstrapServers = append(bootstrapServers, os.Args[1:]...)
	}

	me := node.NewNode(bootstrapServers)
	httpViewer := viewer.NewHttpViewer(me)

	go httpViewer.Serve()
	me.Serve()
}
