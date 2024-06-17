package node

import (
	"context"
	"fmt"
	"gossip-protocol/network"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeSubscription(t *testing.T) {
	logger := NewMockLogger("debug")

	ctx, _ := context.WithCancel(context.Background())
	prvCrypto, _ := ToLibp2pPrivateKey("b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8")

	peer1Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50001))
	peer2Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50002))
	bootsrapAddress, _ := ToAddrInfo([]string{"/ip4/127.0.0.1/tcp/10015/p2p/16Uiu2HAmEH6k5wsPMMSe2Z2vyyeHn8fwPJBL7UDGt4oYCgqPELms"})
	newConfig := Config{
		PrivateKey:      prvCrypto,
		ListenAddresses: []ma.Multiaddr{peer1Address, peer2Address},
		BootstrapPeers:  bootsrapAddress,
	}
	node, err := NewNode(ctx, logger.WithField("layer", "p2p"), newConfig)
	assert.NoError(t, err)
	id := node.ID()
	assert.NotEqual(t, id.String(), "")

	err = node.JoinTopic("go-sammer")
	assert.NoError(t, err)

	addresses := node.ListenAddrs()

	assert.Equal(t, addresses, newConfig.ListenAddresses)
	node.SetUpPeerDiscovery("go-sammer")
	ctx.Done()

}

func TestNewNodeShouldPublishMessage(t *testing.T) {
	logger := NewMockLogger("debug")

	ctx := context.Background()
	prvCrypto, _ := ToLibp2pPrivateKey("b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8")

	peer1Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50001))
	peer2Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50002))
	bootsrapAddress, _ := ToAddrInfo([]string{"/ip4/127.0.0.1/tcp/10015/p2p/16Uiu2HAmEH6k5wsPMMSe2Z2vyyeHn8fwPJBL7UDGt4oYCgqPELms"})
	newConfig := Config{
		PrivateKey:      prvCrypto,
		ListenAddresses: []ma.Multiaddr{peer1Address, peer2Address},
		BootstrapPeers:  bootsrapAddress,
	}
	node, err := NewNode(ctx, logger.WithField("layer", "p2p"), newConfig)
	assert.NoError(t, err)
	id := node.ID()
	assert.NotEqual(t, id.String(), "")

	err = node.JoinTopic("go-sammer")
	assert.NoError(t, err)

	msg := network.Message{
		ID:     "SOMEid0",
		PeerID: "16Uiu2HAmEH6k5wsPMMSe2Z2vyyeHn8fwPJBL7UDGt4oYCgqPELms",
		Data:   []byte{},
	}

	err = node.PublishMessage("go-sammer", msg)
	assert.NoError(t, err)

	node.ConnectToBootstrapPeers()
}

func TestTopicSubscription(t *testing.T) {
	logger := NewMockLogger("debug")

	ctx := context.Background()
	prvCrypto, _ := ToLibp2pPrivateKey("b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8")

	peer1Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50001))
	peer2Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50002))
	bootsrapAddress, _ := ToAddrInfo([]string{"/ip4/127.0.0.1/tcp/10015/p2p/16Uiu2HAmEH6k5wsPMMSe2Z2vyyeHn8fwPJBL7UDGt4oYCgqPELms"})
	newConfig := Config{
		PrivateKey:      prvCrypto,
		ListenAddresses: []ma.Multiaddr{peer1Address, peer2Address},
		BootstrapPeers:  bootsrapAddress,
	}
	node, err := NewNode(ctx, logger.WithField("layer", "p2p"), newConfig)
	assert.NoError(t, err)
	err = node.JoinTopic("topic-1")
	assert.NoError(t, err)
	channel := node.Pubsub.TopicSubscriptionChannel("topic-1")
	assert.NotEqual(t, channel, nil)
	err = node.JoinTopic("topic-2")
	assert.NoError(t, err)

	node.Pubsub.CancelSubscriptions()
	peers := node.Pubsub.GetPeersConnectedToTopic("topic")

	assert.Equal(t, peers, []peer.ID{})

}
