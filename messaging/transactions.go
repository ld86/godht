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
	messaging.transactionMutex.Lock()
	defer messaging.transactionMutex.Unlock()

	messaging.transactionReceivers[id] = receiver
}

func (messaging *Messaging) GetTransactionReceiver(id types.TransactionID) (chan Message, bool) {
	messaging.transactionMutex.Lock()
	defer messaging.transactionMutex.Unlock()

	message, found := messaging.transactionReceivers[id]
	return message, found
}

func (messaging *Messaging) RemoveTransactionReceiver(id types.TransactionID) {
	messaging.transactionMutex.Lock()
	defer messaging.transactionMutex.Unlock()

	delete(messaging.transactionReceivers, id)
}
