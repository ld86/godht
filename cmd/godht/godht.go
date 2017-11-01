package main

import (
    "github.com/ld86/godht/node"
)

func main() {
    me := node.NewNode()
	me.Serve()
}
