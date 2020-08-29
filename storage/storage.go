package storage

import (
	"log"

	badger "github.com/dgraph-io/badger"
)

type Storage struct {
	db *badger.DB

	Messages chan Message
}

type SetKeyRequest struct {
	Key   []byte
	Value []byte
}

type SetKeyResponse struct {
	Err error
}

type GetKeyRequest struct {
	Key []byte
}

type GetKeyResponse struct {
	Value []byte
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
			}
		}
	}
}

func (storage *Storage) SetKey(key []byte, value []byte) error {
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
	message.Response <- response
}

func (storage *Storage) innerSetKey(key []byte, value []byte) error {
	txn := storage.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key, value)
	if err != nil {
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (storage *Storage) GetKey(key []byte) ([]byte, error) {
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

func (storage *Storage) innerGetKey(key []byte) ([]byte, error) {
	txn := storage.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key)

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
