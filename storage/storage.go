package storage

import (
	"errors"
	"log"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/ld86/godht/types"
	"github.com/ld86/godht/utils"
)

type Storage struct {
	db  *badger.DB
	lru *utils.LRU

	Messages chan Message
}

type SetKeyRequest struct {
	Key   types.NodeID
	Value []byte
}

type SetKeyResponse struct {
	Err error
}

type GetKeyRequest struct {
	Key types.NodeID
}

type GetKeyResponse struct {
	Value []byte
	Err   error
}

type OldestElementRequest struct {
}

type OldestElementResponse struct {
	Key   types.NodeID
	Value []byte
	Time  time.Time
	Err   error
}

type Message struct {
	Action   string
	Request  interface{}
	Response chan interface{}
}

func NewStorage() *Storage {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))

	if err != nil {
		log.Fatal(err)
	}

	return &Storage{
		db:       db,
		lru:      utils.NewLRU(),
		Messages: make(chan Message),
	}
}

func (storage *Storage) Serve() {
	storage.handleMessages()
}

func (storage *Storage) handleMessages() {
	for {
		select {
		case message := <-storage.Messages:
			switch message.Action {
			case "SetKey":
				storage.handleSetKey(message)
			case "GetKey":
				storage.handleGetKey(message)
			case "OldestElement":
				storage.handleOldestElement(message)
			}
		}
	}
}

func (storage *Storage) SetKey(key types.NodeID, value []byte) error {
	responseChan := make(chan interface{})
	defer close(responseChan)

	request := SetKeyRequest{
		Key:   key,
		Value: value,
	}
	message := Message{
		Action:   "SetKey",
		Request:  request,
		Response: responseChan,
	}

	storage.Messages <- message
	response := (<-responseChan).(SetKeyResponse)
	return response.Err
}

func (storage *Storage) handleSetKey(message Message) {
	request, ok := message.Request.(SetKeyRequest)
	if !ok {
		return
	}
	err := storage.innerSetKey(request.Key, request.Value)
	response := SetKeyResponse{
		Err: err,
	}

	storage.lru.Store(request.Key, request.Value)

	message.Response <- response
}

func (storage *Storage) innerSetKey(key types.NodeID, value []byte) error {
	txn := storage.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key[:], value)
	if err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) GetKey(key types.NodeID) ([]byte, error) {
	responseChan := make(chan interface{})
	defer close(responseChan)

	request := GetKeyRequest{
		Key: key,
	}
	message := Message{
		Action:   "GetKey",
		Request:  request,
		Response: responseChan,
	}

	storage.Messages <- message
	response := (<-responseChan).(GetKeyResponse)
	return response.Value, response.Err
}

func (storage *Storage) handleGetKey(message Message) {
	request, ok := message.Request.(GetKeyRequest)
	if !ok {
		return
	}
	value, err := storage.innerGetKey(request.Key)
	response := GetKeyResponse{
		Value: value,
		Err:   err,
	}
	message.Response <- response
}

func (storage *Storage) innerGetKey(key types.NodeID) ([]byte, error) {
	txn := storage.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key[:])

	if err != nil {
		return nil, err
	}

	var result []byte

	err = item.Value(func(value []byte) error {
		result = append([]byte{}, value...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (storage *Storage) OldestElement() (types.NodeID, []byte, time.Time, error) {
	responseChan := make(chan interface{})
	defer close(responseChan)

	request := OldestElementRequest{}
	message := Message{
		Action:   "OldestElement",
		Request:  request,
		Response: responseChan,
	}

	storage.Messages <- message
	response := (<-responseChan).(OldestElementResponse)
	return response.Key, response.Value, response.Time, response.Err
}

func (storage *Storage) handleOldestElement(message Message) {
	_, ok := message.Request.(OldestElementRequest)
	if !ok {
		return
	}
	key, value, time, err := storage.innerOldestElement()
	response := OldestElementResponse{
		Key:   key,
		Value: value,
		Time:  time,
		Err:   err,
	}
	message.Response <- response
}

func (storage *Storage) innerOldestElement() (types.NodeID, []byte, time.Time, error) {
	lruElement := storage.lru.OldestElement()

	if lruElement == nil {
		return types.NodeID{}, nil, time.Now(), errors.New("Storage is empty")
	}

	return lruElement.Key.(types.NodeID), lruElement.Value.([]byte), lruElement.Time, nil
}
