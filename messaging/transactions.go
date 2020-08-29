package messaging

import "github.com/ld86/godht/types"

type Transaction struct {
	messasing *Messaging
	id        types.TransactionID

	receiver       chan Message
	messagesToSend chan Message
}

func (messaging *Messaging) NewTransaction() *Transaction {
	transaction := &Transaction{messasing: messaging,
		id: types.NewTransactionID(),

		receiver:       make(chan Message, 100),
		messagesToSend: messaging.messagesToSend,
	}
	messaging.AddTransactionReceiver(transaction.id, transaction.receiver)
	return transaction
}

func (transaction *Transaction) Close() {
	transaction.messasing.RemoveTransactionReceiver(transaction.id)
}

func (transaction *Transaction) SendMessage(message Message) {
	message.TransactionID = &transaction.id
	transaction.messagesToSend <- message
}

func (transaction *Transaction) Receiver() chan Message {
	return transaction.receiver
}

func (messaging *Messaging) AddTransactionReceiver(id types.TransactionID, receiver chan Message) {
	messaging.transactionReceivers.Store(id, receiver)
}

func (messaging *Messaging) GetTransactionReceiver(id types.TransactionID) (chan Message, bool) {
	var result chan Message
	message, found := messaging.transactionReceivers.Load(id)
	if !found {
		result = nil
	} else {
		result = message.(chan Message)
	}
	return result, found
}

func (messaging *Messaging) RemoveTransactionReceiver(id types.TransactionID) {
	messaging.transactionReceivers.Delete(id)
}
