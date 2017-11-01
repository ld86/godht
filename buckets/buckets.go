package buckets

import "container/list"

type Buckets struct {
    k uint
    buckets [160]list.List
}

func NewBuckets(k uint) Buckets {
    return Buckets{k: k}
}
