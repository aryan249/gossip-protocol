package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gossip-protocol/network"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	p2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Node struct {
	ctx    context.Context
	log    *logrus.Entry
	conf   Config
	Host   host.Host
	conn   *ConnectionManager
	Pubsub *PubsubManager
}

func NewNode(ctx context.Context, log *logrus.Entry, conf Config) (*Node, error) {

	connGater := NewConnectionGater(ctx, log.WithField("service", "connection_gater"))
	// create a new libp2p host instance
	h, err := libp2p.New(
		libp2p.Identity(conf.PrivateKey),
		libp2p.ListenAddrs(conf.ListenAddresses...),
		libp2p.Ping(true),
		libp2p.WithDialTimeout(5*time.Second),
		libp2p.ConnectionGater(connGater),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Add bootstrap nodes permanently to node's peer store
	for _, bootstrapNode := range conf.BootstrapPeers {
		h.Peerstore().AddAddrs(bootstrapNode.ID, bootstrapNode.Addrs, peerstore.PermanentAddrTTL)
	}

	// Create a new connection manager service instance
	conn := NewConnectionManager(ctx, log.WithField("service", "connection_manager"), h)

	// Create a new pubsub manager service instance
	pubsub, err := NewPubsubManager(ctx, log.WithField("service", "pubsub_manager"), h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub manager: %w", err)
	}

	node := &Node{
		ctx:    ctx,
		log:    log,
		conf:   conf,
		Host:   h,
		conn:   conn,
		Pubsub: pubsub,
	}

	return node, nil
}

// Initialize a Decentralized Hash Table, following the Kademlia implementation
func (n *Node) NewDHT() (*dht.IpfsDHT, error) {
	var options []dht.Option

	// Currently every node is added as a server mode, but we may want to configure this
	options = append(options, dht.Mode(dht.ModeServer))

	kdht, err := dht.New(n.ctx, n.Host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(n.ctx); err != nil {
		return nil, err
	}

	return kdht, nil
}

func (n *Node) ID() peer.ID {
	return n.Host.ID()
}

func (n *Node) ListenAddrs() []ma.Multiaddr {
	return n.Host.Addrs()
}

func (n *Node) Peerstore() peerstore.Peerstore {
	return n.Host.Peerstore()
}

func (n *Node) Network() p2pNetwork.Network {
	return n.Host.Network()
}

// Initiates peer discovery on the given topic, based on relevant discovery methods configured for the node
func (n *Node) SetUpPeerDiscovery(topic string) {

	if slices.Contains(n.conf.DiscoveryMethods, DiscoveryMethodDHT) {
		n.log.Info("DHT peer discovery enabled for node")
		dht, _ := n.NewDHT()
		go DiscoverDHT(n.ctx, topic, dht, n.Host, n.log.Logger)
	}

}

// ConnectToBootstrapPeers connects to the bootstrap peers.
func (n *Node) ConnectToBootstrapPeers() {
	var wg sync.WaitGroup
	for _, peerInfo := range n.conf.BootstrapPeers {
		wg.Add(1)

		info := peerInfo // copy to avoid race conditions
		logPeer := n.log.WithField("peer", info)

		logPeer.Debug("Connecting to bootstrap node")
		go func() {
			defer wg.Done()

			timeoutCtx, cancel := context.WithTimeout(n.ctx, 5*time.Second)
			defer cancel()

			if err := n.Host.Connect(timeoutCtx, *info); err != nil {
				logPeer.Warnf("Error while connecting to peer: %-v", err)
				// We don't want to remove the bootstrap peer from the peer store.
				// Since bootstrap peers are expected to be generally available,
				// we want the connection handler to retry the connection later,
				// opposed to a regular peer that may just have gone offline
				return
			}

			logPeer.Info("Connection established with bootstrap peer")
		}()
	}

	wg.Wait()
}

func (n *Node) JoinTopic(topic string) error {
	return n.Pubsub.JoinTopic(topic)
}

func (n *Node) PublishMessage(topic string, msg network.Message) error {
	return n.Pubsub.PublishMessage(topic, msg)
}

func (n *Node) SubscriptionChannel(topic string) <-chan network.Message {
	return n.Pubsub.TopicSubscriptionChannel(topic)
}

// Start starts the node subscriptions read loop.
func (n *Node) Start(wg *sync.WaitGroup) {
	n.Pubsub.Start(wg)

	wg.Add(1)
	go n.conn.Start(wg)
}

func (n *Node) Stop() {
	n.Pubsub.CancelSubscriptions()

	// Close the libp2p host
	if err := n.Host.Close(); err != nil {
		n.log.WithError(err).Error("failed to close libp2p host")
	}
}
