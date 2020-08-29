package messaging

import (
	"encoding/json"
	"log"
	"net"
)

func (messaging *Messaging) handleSentMessages() {
	for {
		select {
		case outputMessage := <-messaging.messagesToSend:
			if outputMessage.FromId == outputMessage.ToId {
				log.Printf("Drop message to yourself")
				continue
			}
			var remoteAddr net.Addr

			if outputMessage.IpAddr == nil {
				var ok bool
				messaging.mappingMutex.Lock()
				remoteAddr, ok = messaging.mapping[outputMessage.ToId]
				messaging.mappingMutex.Unlock()
				if !ok {
					log.Printf("Cannot find remote addr for node with id %s, skipping message", outputMessage.ToId)
					continue
				}
				log.Printf("Found remoteAddr %s by id %s", remoteAddr, outputMessage.ToId.String())
			} else {
				var err error
				remoteAddr, err = net.ResolveUDPAddr("udp", *outputMessage.IpAddr)
				if err != nil {
					log.Printf("Cannot resolve %s, %s", *outputMessage.IpAddr, err)
					continue
				}
				log.Printf("Resolved remoteAddr %s", *outputMessage.IpAddr)
			}

			{
				messaging.mappingMutex.Lock()
				defer messaging.mappingMutex.Unlock()
				outputMessage.IdToAddrMapping = make([]IdAddr, 0)
				for _, nodeID := range outputMessage.Ids {
					nodeAddr, ok := messaging.mapping[nodeID]
					if !ok {
						continue
					}
					outputMessage.IdToAddrMapping = append(outputMessage.IdToAddrMapping,
						IdAddr{nodeID, nodeAddr.String()})
				}
			}

			log.Printf("Trying to send message %s", outputMessage.String())
			data, _ := json.Marshal(outputMessage)
			messaging.serverConnection.WriteTo(data, remoteAddr)
		}
	}
}
