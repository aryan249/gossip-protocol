package p2p

import (
	"sync"

	"gossip-protocol/network"
)

var lock = sync.Mutex{}

// signAndPublishRelayersMessage sends a message to the relevant pubsub topic.
func (n *Node) PublishRelayersMessage(topic string, addressedMsg network.Message) error {
	lock.Lock()
	defer lock.Unlock()

	if err := n.Node.PublishMessage(topic, addressedMsg); err != nil {
		return err
	}

	return nil
}

func (n *Node) PublishMessageLoop(topic string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case msg, ok := <-n.obsSendReq:
			if !ok {
				n.log.Errorf("error in publishing")
				continue
			}

			if err := n.PublishRelayersMessage(topic, msg); err != nil {
				n.log.WithError(err).Errorf("publish error")
			}
		case <-n.ctx.Done():
			return
		}
	}
}

type NewMessageHandler struct {
	n *Node
}

func (r *NewMessageHandler) newMessage() network.Message {
	return network.Message{}
}

// MessageLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (mh *NewMessageHandler) MessageLoop(topic string, wg *sync.WaitGroup) {
	defer wg.Done()

	channel := mh.n.Node.SubscriptionChannel(topic)
	if channel == nil {
		// This means the topic has been unsubscribed
		return
	}

	for {
		select {
		case <-mh.n.ctx.Done():
			return
		case cm, ok := <-channel:
			if !ok {
				return
			}

			mh.n.obsReceiveRes <- cm

		}
	}
}
