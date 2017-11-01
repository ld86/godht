package main

import (
    "fmt"
    "github.com/ld86/godht/buckets"
)

func main() {
    buckets := buckets.NewBuckets(10)
    fmt.Println(buckets)
}
