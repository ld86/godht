package utils

import (
	"container/list"
	"sync"
	"time"
)

type LRU struct {
	m *sync.Map
	l *list.List

	mutex *sync.Mutex
}

type LRUElement struct {
	Key   interface{}
	Value interface{}
	Time  time.Time
}

func NewLRU() *LRU {
	return &LRU{m: &sync.Map{},
		l:     &list.List{},
		mutex: &sync.Mutex{},
	}
}

func (lru *LRU) Store(k, v interface{}) bool {
	listElement, found := lru.m.Load(k)

	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if found {
		lru.l.MoveToBack(listElement.(*list.Element))
		listElement.(*list.Element).Value.(*LRUElement).Time = time.Now()
		return false
	}

	newListElement := &LRUElement{Key: k, Value: v, Time: time.Now()}
	lru.m.Store(k, lru.l.PushBack(newListElement))

	return true
}

func (lru *LRU) Load(k interface{}) (interface{}, bool) {
	listElement, found := lru.m.Load(k)
	if !found {
		return nil, false
	}
	return listElement.(*list.Element).Value.(*LRUElement).Value, true
}

func (lru *LRU) Delete(k interface{}) bool {
	listElement, found := lru.m.Load(k)
	if !found {
		return false
	}

	lru.mutex.Lock()
	lru.l.Remove(listElement.(*list.Element))
	lru.mutex.Unlock()
	lru.m.Delete(k)

	return true
}

func (lru *LRU) Touch(k interface{}) bool {
	listElement, found := lru.m.Load(k)
	if found {
		lru.mutex.Lock()
		defer lru.mutex.Unlock()

		lru.l.MoveToBack(listElement.(*list.Element))
		listElement.(*list.Element).Value.(*LRUElement).Time = time.Now()
		return true
	}
	return false
}

func (lru *LRU) LastTouched() *LRUElement {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if lru.l.Len() == 0 {
		return nil
	}

	return lru.l.Back().Value.(*LRUElement)
}
