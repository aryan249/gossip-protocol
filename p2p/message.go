package p2p

import (
	"sync"

	"gossip-protocol/network"
)

type MessageHandler interface {
	newMessage() network.Message
	MessageLoop(string, *sync.WaitGroup)
}
