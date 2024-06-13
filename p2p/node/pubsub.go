package node

import (
	"context"
	"encoding/json"
	"fmt"
	"gossip-protocol/network"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

const DefaultTopicBufferSize = 1024

type TopicSubscription struct {
	topicHandle              *pubsub.Topic
	topicSubscription        *pubsub.Subscription
	topicSubscriptionChannel chan network.Message
	newMessage               func() network.Message
}

// PubsubManager abstracts away resource management related to topic subscription handling.
// It exposes methods to join and publish messages to arbitrary topics
type PubsubManager struct {
	ctx context.Context
	log *logrus.Entry

	host   host.Host
	pubSub *pubsub.PubSub

	topicSubscriptions map[string]*TopicSubscription
	mapLock            sync.Mutex
}

func NewPubsubManager(ctx context.Context, logger *logrus.Entry, host host.Host) (*PubsubManager, error) {
	// Create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, host, pubsub.WithValidateQueueSize(50000))
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub service: %w", err)
	}

	manager := &PubsubManager{
		ctx:                ctx,
		log:                logger.WithField("service", "pubsub_manager"),
		host:               host,
		pubSub:             ps,
		topicSubscriptions: map[string]*TopicSubscription{},
	}

	return manager, nil
}

// Join a given topic.
func (pm *PubsubManager) JoinTopic(topic string) error {

	pm.mapLock.Lock()
	defer pm.mapLock.Unlock()

	pm.log.Debugf("Joining topic %s", topic)

	// idempotency check: if the topic has already been joined, there is no need to perform any action
	if _, containsKey := pm.topicSubscriptions[topic]; !containsKey {

		// Join the topic
		topicHandle, err := pm.pubSub.Join(topic)
		if err != nil {
			return fmt.Errorf("failed to join '%s': %w", topic, err)
		}

		// Subscribe to the topic
		topicSubscription, err := topicHandle.Subscribe(pubsub.WithBufferSize(10000))
		if err != nil {
			return fmt.Errorf("failed to subscribe to '%s': %w", topic, err)
		}

		// Create a channel to read messages from the topic
		topicSubscriptionChannel := make(chan network.Message, DefaultTopicBufferSize)

		pm.topicSubscriptions[topic] = &TopicSubscription{
			topicHandle:              topicHandle,
			topicSubscription:        topicSubscription,
			topicSubscriptionChannel: topicSubscriptionChannel,
		}
	}

	return nil
}

func (pm *PubsubManager) GetPeersConnectedToTopic(topic string) []peer.ID {
	if _, containsKey := pm.topicSubscriptions[topic]; containsKey {
		return pm.topicSubscriptions[topic].topicHandle.ListPeers()
	} else {
		return []peer.ID{}
	}
}

func (pm *PubsubManager) PublishMessage(topic string, message network.Message) error {

	pm.mapLock.Lock()
	defer pm.mapLock.Unlock()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	if _, containsKey := pm.topicSubscriptions[topic]; containsKey {
		// If we are already subscribed to the topic
		return pm.topicSubscriptions[topic].topicHandle.Publish(pm.ctx, data)
	} else {
		// Otherwise, this is a a one-time use topic.
		// Join the topic, Publish the message and unsuscribe, to avoid accumulating resources
		pm.log.Debugf("Joining topic %s", topic)
		peerTopic, err := pm.pubSub.Join(topic)
		if err != nil {
			return fmt.Errorf("failed to join topic '%s': %w", topic, err)
		}
		defer peerTopic.Close()
		return peerTopic.Publish(pm.ctx, data)
	}

}

func (pm *PubsubManager) TopicSubscriptionChannel(topic string) <-chan network.Message {

	pm.mapLock.Lock()
	defer pm.mapLock.Unlock()

	if _, containsKey := pm.topicSubscriptions[topic]; containsKey {
		return pm.topicSubscriptions[topic].topicSubscriptionChannel
	} else {
		return nil
	}
}

func (pm *PubsubManager) cancelTopicSubscription(topic string) {

	pm.mapLock.Lock()
	defer pm.mapLock.Unlock()

	if _, containsKey := pm.topicSubscriptions[topic]; containsKey {

		// We do not need to close the channel itself.
		// As mentioned in go documentation: [closing the channel] should be executed only by the sender, never the receiver,

		// Cancel the topic subscription
		pm.topicSubscriptions[topic].topicSubscription.Cancel()

		// Close the topic handle
		if err := pm.topicSubscriptions[topic].topicHandle.Close(); err != nil {
			pm.log.WithError(err).Errorf("failed to close topic %s", topic)
		}

		delete(pm.topicSubscriptions, topic)
	}
}

func (pm *PubsubManager) processTopic(topic string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Loop, which will continue as long as we are still subscribed
	for {

		// first, determine if we are subscribed, unlocking right away once we make the determination
		pm.mapLock.Lock()

		var subscription *pubsub.Subscription

		if _, containsKey := pm.topicSubscriptions[topic]; containsKey {
			subscription = pm.topicSubscriptions[topic].topicSubscription
		}

		pm.mapLock.Unlock()

		// if we are not subscribed to the topic, there is nothing to process anymore.
		if subscription == nil {
			return
		}

		// wait for the next message. This is a blocking operation and it is possible that, as we wait,
		// the topic becomes unsubscribed.
		msg, err := subscription.Next(pm.ctx)

		if err != nil {
			// if message cannot be read, ensure the topic is cleanly closed.
			// Based on the implementation of libp2p, we should really only
			// get an error in the topic was unsubscribed. This means two things:
			// - Calling cancelTopicSubscription here is most of the time unnecessary. However, there is no harm in doing it, as it is idempotent and would result in no extra operation
			// - We are not going to try to recover from the error. The error appears because another goroutine willingly unsubscribed.
			pm.cancelTopicSubscription(topic)
			pm.log.WithError(err).Warn("failed to read message from topic")
			return
		}

		// Only process messages delivered by others
		if msg.ReceivedFrom == pm.host.ID() {
			continue
		}

		// Now we have a valid message. Between the time we started listening for the next message, and the time we actually received this message, the topic could have been unsubscribed from.
		// Thus, we need to perform a second check
		pm.mapLock.Lock()

		// if we are still interested in messages from this topic

		var bm network.Message
		if _, containsKey := pm.topicSubscriptions[topic]; containsKey {
			bm = network.Message{}
			err = json.Unmarshal(msg.Data, bm)
			if err != nil {
				pm.log.WithError(err).Trace("failed to unmarshal message")
				pm.mapLock.Unlock()
				continue
			}
		}
		pm.mapLock.Unlock()

		// Send the message to the topic channel.
		// This is done after unlocking the lock, to ensure there is no deadlocking
		// between mapLock, and the topicSubscriptionChannel's internal lock
		if bm.Data != nil {
			pm.topicSubscriptions[topic].topicSubscriptionChannel <- bm
		}
	}
}

func (pm *PubsubManager) CancelSubscriptions() {

	topics := []string{}
	pm.mapLock.Lock()
	for topic := range pm.topicSubscriptions {
		topics = append(topics, topic)
	}
	pm.mapLock.Unlock()

	for topic := range pm.topicSubscriptions {
		pm.cancelTopicSubscription(topic)
	}
}

func (pm *PubsubManager) Start(wg *sync.WaitGroup) {
	for topic := range pm.topicSubscriptions {
		wg.Add(1)
		go pm.processTopic(topic, wg)
	}
}
