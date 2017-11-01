package main

import (
    "fmt"
    _ "github.com/ld86/godht/buckets"
    "github.com/ld86/godht/node"
)

func main() {
    // buckets := buckets.NewBuckets(10)
    firstNode := node.NewNode()
    secondNode := node.NewNode()
    fmt.Println(firstNode.Distance(secondNode))
    fmt.Println(firstNode.GetBucketIndex(secondNode))
}
