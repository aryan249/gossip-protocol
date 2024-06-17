package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"gossip-protocol/config"
	"gossip-protocol/network"
	"gossip-protocol/p2p/node"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

// This utility will join same perp api topic as stated in constants and same topic willl eb joined by relayers
// this will run as publisher node and relayers will be subscribers
// Notes: This utility can be used to test orders insertion via p2p so that this node while running can publish order in every 10 seconds that will be created
// - createMockP2pOrder method can be used for order creation ans send orders go routine publishes order messages
func main() {

	flag.Parse()

	ctx := context.Background()
	// create new logger
	logger := NewRootLogger("debug")

	// read config from config files in config files directory in perp api insert order directory
	cfg := config.NewViperConfig()
	loggerLevel := cfg.ReadLogLevelConfig()

	// configure logger
	logLevel, err := logrus.ParseLevel(loggerLevel)
	if err == nil {
		logger.SetLevel(logLevel)
	}

	prvCrypto, _ := node.ToLibp2pPrivateKey("b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8") // enter private key for perp api address that is set in relayers or can be set in config
	prvEcdsaKey, _ := eth_crypto.HexToECDSA("b2eada77569027933f70e63fbb356a1193bcb3c3d52cba3859d5fd73103d69c8") // enter private key for perp api address that is set in relayers

	// Retrieve target topic to connect to
	syncTopic := "go-sammer"

	// Parse bootstrap peer. This is the peer that we expect to be listening to the perp api sync topic
	ma, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10015/p2p/16Uiu2HAmEH6k5wsPMMSe2Z2vyyeHn8fwPJBL7UDGt4oYCgqPELms")
	bootstrapPeer, _ := peer.AddrInfoFromP2pAddr(ma)

	addresses := []string{"/ip4/0.0.0.0/tcp/10016", "/ip6/::/tcp/10016"}

	addrs := make([]multiaddr.Multiaddr, 0, len(addresses))
	for _, addr := range addresses {
		ma, _ := multiaddr.NewMultiaddr(addr)

		addrs = append(addrs, ma)
	}

	// new host

	h, _ := libp2p.New(
		libp2p.Identity(prvCrypto),
		libp2p.ListenAddrs(addrs...),
		libp2p.Ping(true),
		libp2p.WithDialTimeout(5*time.Second),
	)

	// Add bootstrap nodes permanently to node's peer store
	h.Peerstore().AddAddrs(bootstrapPeer.ID, bootstrapPeer.Addrs, peerstore.PermanentAddrTTL)

	var wg sync.WaitGroup

	// Create a new connection manager service instance
	conn := node.NewConnectionManager(ctx, logger.WithField("service", "connection_manager"), h)
	wg.Add(1)
	go conn.Start(&wg)

	// create a new pubsub manager
	pubsub, err := node.NewPubsubManager(ctx, logger.WithField("service", "pubsub_manager"), h)

	err = pubsub.JoinTopic(syncTopic)
	if err == nil {
		wg.Add(1)
		pubsub.Start(&wg)

		wg.Add(1)
		// go routine to publish messages
		go sendOrders(syncTopic, prvEcdsaKey, pubsub, h, logger, &wg)

		wg.Wait()
	} else {
		logger.Errorf("Could not join topic %s", syncTopic)
	}
}

func sendOrders(syncTopic string, prvEcdsaKey *ecdsa.PrivateKey, pubsub *node.PubsubManager, h host.Host, logger *logrus.Logger, wg *sync.WaitGroup) {
	time.Sleep(65 * time.Second)
	for {
		newMessage := network.Message{
			ID:     "someID1",
			PeerID: h.ID().String(),
			Data:   []byte{},
		}

		logger.Infof("publishing order request to %s", syncTopic)
		err := pubsub.PublishMessage(syncTopic, newMessage)
		if err != nil {
			// Log about error, and hope we can recover next loop.
			// This is probably a temporary network disconnect
			logger.Errorf("when publishing orders %s", syncTopic, err)
		}
		time.Sleep(10 * time.Second)
	}
	time.Sleep(30 * time.Second)
	os.Exit(0)
}

func NewRootLogger(loggerLevel string) *logrus.Logger {
	// Initialize root logger
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
