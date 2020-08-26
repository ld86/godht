package messaging

import "github.com/ld86/godht/types"

type Transaction struct {
	messasing *Messaging
	id        types.TransactionID

	Receiver       chan Message
	MessagesToSend chan Message
}

func (messaging *Messaging) NewTransaction() *Transaction {
	transaction := &Transaction{messasing: messaging,
		id: types.NewTransactionID(),

		Receiver:       make(chan Message),
		MessagesToSend: messaging.messagesToSend,
	}
	return transaction
}

func (messaging *Messaging) AddTransactionReceiver(id types.TransactionID, receiver chan Message) {
	messaging.mutex.Lock()
	defer messaging.mutex.Unlock()

	messaging.transactionReceivers[id] = receiver
}

func (messaging *Messaging) GetTransactionReceiver(id types.TransactionID) (chan Message, bool) {
	messaging.mutex.Lock()
	defer messaging.mutex.Unlock()

	message, found := messaging.transactionReceivers[id]
	return message, found
}

func (messaging *Messaging) RemoveTransactionReceiver(id types.TransactionID) {
	messaging.mutex.Lock()
	defer messaging.mutex.Unlock()

	delete(messaging.transactionReceivers, id)
}
