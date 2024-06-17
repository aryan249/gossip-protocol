package p2p

import (
	"context"
	"fmt"
	"gossip-protocol/network"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	obsReceiveRes := make(chan network.Message, 1)
	obsSendReq := make(chan network.Message, 1)

	logger := NewMockLogger("debug")

	ctx, _ := context.WithCancel(context.Background())

	config := Config{
		ListenAddresses:  []string{"/ip4/0.0.0.0/tcp/10016"},
		PrivateKey:       "b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8",
		DiscoveryMethods: []string{"mdns"},
	}

	newNode := NewP2PNode(ctx, logger.WithField("layer", "p2p"), obsReceiveRes, obsSendReq, config)

	assert.NotEmpty(t, newNode)

	peerId := newNode.GetPeerID()
	assert.NotEqual(t, peerId, "")

	address := newNode.GetAddress()
	assert.Equal(t, address, "09426b744f04cdce9af3e1df5acea60f3d4aadb5")
}

func TestNewNodeGetMessage(t *testing.T) {
	var wg sync.WaitGroup
	obsReceiveRes := make(chan network.Message, 1)
	obsSendReq := make(chan network.Message, 1)

	logger := NewMockLogger("debug")

	ctx, _ := context.WithCancel(context.Background())

	config := Config{
		ListenAddresses:  []string{"/ip4/0.0.0.0/tcp/10016"},
		PrivateKey:       "b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8",
		DiscoveryMethods: []string{"mdns"},
	}

	newNode := NewP2PNode(ctx, logger.WithField("layer", "p2p"), obsReceiveRes, obsSendReq, config)

	assert.NotEmpty(t, newNode)

	peerId := newNode.GetPeerID()
	assert.NotEqual(t, peerId, "")

	address := newNode.GetAddress()
	assert.Equal(t, address, "09426b744f04cdce9af3e1df5acea60f3d4aadb5")
	wg.Add(1)
	newNode.Start(&wg)
	channel1 := newNode.Node.SubscriptionChannel("topic-1")
	assert.NotEqual(t, channel1, nil)
	channel := newNode.Node.Pubsub.GetTopicSubscriptionChannel("go-sammer")

	newMesage := network.Message{
		ID:     "SomeId1",
		PeerID: newNode.GetPeerID().String(),
		Data:   []byte{},
	}

	channel <- newMesage

	msg, ok := <-obsReceiveRes
	assert.Equal(t, ok, true)
	assert.Equal(t, msg.ID, newMesage.ID)
	ctx.Done()

}

func TestReturnFalseForInvalidPeerConnection(t *testing.T) {
	obsReceiveRes := make(chan network.Message, 1)
	obsSendReq := make(chan network.Message, 1)

	logger := NewMockLogger("debug")

	ctx, _ := context.WithCancel(context.Background())

	config := Config{
		ListenAddresses:  []string{"/ip4/0.0.0.0/tcp/10016"},
		PrivateKey:       "b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8",
		DiscoveryMethods: []string{"mdns"},
	}

	newNode := NewP2PNode(ctx, logger.WithField("layer", "p2p"), obsReceiveRes, obsSendReq, config)

	result := newNode.IsPeerConnected("78327463723776713613849237483264873279")
	assert.Equal(t, result, false)
	// but since in peer list it has only one peer
	hasPeers := newNode.HasPeers()

	assert.Equal(t, hasPeers, true)

}

func TestShouldPublishMessage(t *testing.T) {
	var wg sync.WaitGroup

	obsReceiveRes := make(chan network.Message, 1)
	obsSendReq := make(chan network.Message, 1)

	logger := NewMockLogger("debug")

	ctx, _ := context.WithCancel(context.Background())

	config := Config{
		ListenAddresses:  []string{"/ip4/0.0.0.0/tcp/10016"},
		PrivateKey:       "b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8",
		DiscoveryMethods: []string{"mdns"},
	}

	newNode := NewP2PNode(ctx, logger.WithField("layer", "p2p"), obsReceiveRes, obsSendReq, config)

	newMesage := network.Message{
		ID:     "SomeId1",
		PeerID: newNode.GetPeerID().String(),
		Data:   []byte{},
	}

	err := newNode.PublishRelayersMessage("go-sammer", newMesage)

	assert.NoError(t, err)

	// message loop will return if there is no topic subscriobed
	handler := NewMessageHandler{
		n: newNode,
	}

	wg.Add(1)

	handler.MessageLoop("topic-1", &wg)
	wg.Wait()
	ctx.Done()
}

func NewMockLogger(loggerLevel string) *logrus.Logger {

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	logLevel, err := logrus.ParseLevel(loggerLevel)
	if err != nil {
		panic(fmt.Errorf("failed to parse log level: %v", err))
	}
	logger.SetLevel(logLevel)

	return logger
}
