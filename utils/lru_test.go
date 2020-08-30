package utils_test

import (
	"testing"
	"time"

	"github.com/ld86/godht/utils"
)

func TestSimpleSetGet(t *testing.T) {
	lru := utils.NewLRU()

	_, found := lru.Load("1")
	if found {
		t.Error("LRU should be empty")
	}

	added := lru.Store("1", "1")
	if !added {
		t.Error("Can't store new element")
	}

	value, found := lru.Load("1")

	if !found {
		t.Error("Value should exist")
	}

	if value.(string) != "1" {
		t.Error("Wrong returned value")
	}
}

func TestLastTouched(t *testing.T) {
	lru := utils.NewLRU()

	if lru.LastTouched() != nil {
		t.Error("LastTouched value of empty LRU should be nil")
	}

	added := lru.Store("1", "1")
	if !added {
		t.Error("Can't store new element")
	}

	lastTouched := lru.LastTouched()

	if lastTouched == nil {
		t.Error("LastTouched value of non-empty LRU should not be nil")
	}

	if lastTouched.Key.(string) != "1" || lastTouched.Value.(string) != "1" {
		t.Error("Wrong last touched element")
	}

	if !lastTouched.Time.Before(time.Now()) {
		t.Error("Time should be less or equal to time.Now()")
	}

	added = lru.Store("2", "2")
	if !added {
		t.Error("Can't store new element")
	}

	lastTouched = lru.LastTouched()

	if lastTouched == nil {
		t.Error("LastTouched value of non-empty LRU should not be nil")
	}

	if lastTouched.Key.(string) != "2" || lastTouched.Value.(string) != "2" {
		t.Error("Wrong last touched element")
	}

	if !lastTouched.Time.Before(time.Now()) {
		t.Error("Time should be less or equal to time.Now()")
	}
}

func TestTouch(t *testing.T) {
	lru := utils.NewLRU()

	touched := lru.Touch("1")
	if touched {
		t.Error("Can't touch element of empty LRU")
	}

	added := lru.Store("1", "1")
	if !added {
		t.Error("Can't store new element")
	}

	added = lru.Store("2", "2")
	if !added {
		t.Error("Can't store new element")
	}

	lastTouched := lru.LastTouched()

	if lastTouched == nil {
		t.Error("LastTouched value of non-empty LRU should not be nil")
	}

	if lastTouched.Key.(string) != "2" || lastTouched.Value.(string) != "2" {
		t.Error("Wrong last touched element")
	}

	touched = lru.Touch("1")
	if !touched {
		t.Error("Touched key presented")
	}

	lastTouched = lru.LastTouched()

	if lastTouched == nil {
		t.Error("LastTouched value of non-empty LRU should not be nil")
	}

	if lastTouched.Key.(string) != "1" || lastTouched.Value.(string) != "1" {
		t.Error("Wrong last touched element")
	}

}
func TestDelete(t *testing.T) {
	lru := utils.NewLRU()

	deleted := lru.Delete("1")
	if deleted {
		t.Error("Non-existing key should not be deleted")
	}

	added := lru.Store("1", "1")
	if !added {
		t.Error("Can't store new element")
	}

	added = lru.Store("2", "2")
	if !added {
		t.Error("Can't store new element")
	}

	added = lru.Store("3", "3")
	if !added {
		t.Error("Can't store new element")
	}

	lastTouched := lru.LastTouched()
	if lastTouched == nil {
		t.Error("LastTouched value of non-empty LRU should not be nil")
	}

	if lastTouched.Key.(string) != "3" || lastTouched.Value.(string) != "3" {
		t.Error("Wrong last touched element")
	}

	deleted = lru.Delete("1")
	if !deleted {
		t.Error("Existing key should be deleted")
	}

	_, found := lru.Load("1")
	if found {
		t.Error("Deleted key should be loaded")
	}

	lastTouched = lru.LastTouched()

	if lastTouched == nil {
		t.Error("LastTouched value of non-empty LRU should not be nil")
	}

	if lastTouched.Key.(string) != "3" || lastTouched.Value.(string) != "3" {
		t.Error("Wrong last touched element")
	}
}
