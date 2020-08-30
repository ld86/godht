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
