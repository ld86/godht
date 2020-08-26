package messaging

import "github.com/ld86/godht/types"

func (messaging *Messaging) AddTransactionReceiver(id types.TransactionID, receiver chan Message) {
	messaging.mutex.Lock()
	defer messaging.mutex.Unlock()

	messaging.TransactionReceivers[id] = receiver
}

func (messaging *Messaging) RemoveTransactionReceiver(id types.TransactionID) {
	messaging.mutex.Lock()
	defer messaging.mutex.Unlock()

	delete(messaging.TransactionReceivers, id)
}
