package messaging

import (
	"encoding/json"
	"log"
	"net"
)

func (messaging *Messaging) handleReceivedMessages() {
	for {
		var buffer [2048]byte
		var message Message
		n, remoteAddr, _ := messaging.serverConnection.ReadFrom(buffer[:])
		json.Unmarshal(buffer[:n], &message)

		log.Printf("Received message %s", message.String())
		log.Printf("Remember node with id %s by remoteAddr %s", message.FromId.String(), remoteAddr)

		messaging.mapping[message.FromId] = remoteAddr
		for _, idAddr := range message.IdToAddrMapping {
			log.Printf("Remember node with id %s by remoteAddr %s", idAddr.Id.String(), idAddr.Addr)
			messaging.mapping[idAddr.Id], _ = net.ResolveUDPAddr("udp", idAddr.Addr)
		}

		if message.TransactionID == nil {
			if messaging.DefaultReceiver != nil {
				messaging.DefaultReceiver <- message
			}
		} else {
			channel, found := messaging.TransactionReceivers[*message.TransactionID]
			if !found {
				channel = messaging.DefaultReceiver
			}
			if channel != nil {
				channel <- message
			}
		}
	}
}
