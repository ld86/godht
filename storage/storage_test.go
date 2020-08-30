package storage_test

import (
	"testing"

	"github.com/ld86/godht/storage"
	"github.com/ld86/godht/types"
)

func TestSimpleSetGet(t *testing.T) {
	storage := storage.NewStorage()
	go storage.Serve()

	for i := 0; i < 100; i++ {
		key := types.NewNodeID()
		value := types.NewNodeID()

		err := storage.SetKey(key, []byte(value.String()))
		if err != nil {
			t.Error(err)
		}

		retrievedValue, err := storage.GetKey(key)

		if err != nil {
			t.Error(err)
		}

		if string(retrievedValue) != value.String() {
			t.Errorf("%s != %s", string(retrievedValue), value.String())
		}
	}
}

func TestSetSame(t *testing.T) {
	storage := storage.NewStorage()
	go storage.Serve()

	key := types.NewNodeID()

	err := storage.SetKey(key, []byte("1"))
	if err != nil {
		t.Error(err)
	}

	value, err := storage.GetKey(key)
	if err != nil {
		t.Error(err)
	}

	if string(value) != "1" {
		t.Errorf("%s != 1", string(value))
	}

	err = storage.SetKey(key, []byte("2"))
	if err != nil {
		t.Error(err)
	}

	value, err = storage.GetKey(key)
	if err != nil {
		t.Error(err)
	}

	if string(value) != "2" {
		t.Errorf("%s != 2", string(value))
	}
}

func TestGetOnly(t *testing.T) {
	storage := storage.NewStorage()
	go storage.Serve()

	key := types.NewNodeID()

	_, err := storage.GetKey(key)
	if err == nil {
		t.Error(err)
	}

}

func TestOldestElement(t *testing.T) {
	storage := storage.NewStorage()
	go storage.Serve()

	keyA := types.NewNodeID()
	keyB := types.NewNodeID()

	_, _, err := storage.OldestElement()
	if err == nil {
		t.Error("Oldest element of empty storage should not be found")
	}

	err = storage.SetKey(keyA, []byte("1"))
	if err != nil {
		t.Error(err)
	}

	oldestKey, oldestValue, err := storage.OldestElement()
	if err != nil {
		t.Error(err)
	}

	if oldestKey != keyA || string(oldestValue) != "1" {
		t.Error("Wrong oldest value")
	}

	err = storage.SetKey(keyB, []byte("2"))
	if err != nil {
		t.Error(err)
	}

	oldestKey, oldestValue, err = storage.OldestElement()
	if err != nil {
		t.Error(err)
	}

	if oldestKey != keyA || string(oldestValue) != "1" {
		t.Error("Wrong oldest value")
	}

	err = storage.SetKey(keyA, []byte("3"))
	if err != nil {
		t.Error(err)
	}

	oldestKey, oldestValue, err = storage.OldestElement()
	if err != nil {
		t.Error(err)
	}

	if oldestKey != keyB || string(oldestValue) != "2" {
		t.Error("Wrong oldest value")
	}

}
