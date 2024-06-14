package p2p

import (
	"context"
	"crypto/ecdsa"
	"sync"

	"gossip-protocol/network"
	"gossip-protocol/p2p/node"

	"github.com/ethereum/go-ethereum/common"
	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/libp2p/go-libp2p/core/crypto"
	p2pNetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

const clientVersion = "perp-relayer/" + "go-scammer"

// Node type - a p2p host implementing one or more p2p protocols
type Node struct {
	ctx  context.Context
	log  *logrus.Entry
	Node *node.Node

	lock            sync.Mutex
	prvSecp256k1Key *crypto.Secp256k1PrivateKey
	prvEcdsaKey     *ecdsa.PrivateKey
	pubKeyHash      string

	emitBeaconMessages bool
	EnableSync         bool

	obsReceiveRes chan<- network.Message
	obsSendReq    <-chan network.Message
	SendToPerpApi bool
}

func NewP2PNode(ctx context.Context, log *logrus.Entry, obsReceiveRes chan<- network.Message, obsSendReq <-chan network.Message, p2pCfg Config) *Node {
	return newNode(
		ctx, log, p2pCfg,
		obsReceiveRes, obsSendReq,
	)
}

// helper method - create a lib-p2p host to listen on a port
func newNode(ctx context.Context, log *logrus.Entry, p2pCfg Config, obsReceiveRes chan<- network.Message, obsSendReq <-chan network.Message) *Node {
	// Parse the p2pNode's private key as a libp2p secp256k1 key
	prvSecp256k1Key, err := crypto.UnmarshalSecp256k1PrivateKey(common.Hex2Bytes(p2pCfg.PrivateKey))
	if err != nil {
		log.Panicf("invalid p2pNode private key: %v", err)
	}

	// Parse the p2pNode's private key as a geth ecdsa key
	prvEcdsaKey, err := eth_crypto.HexToECDSA(p2pCfg.PrivateKey)
	if err != nil {
		log.Panicf("invalid p2pNode private key: %v", err)
	}

	// Compute the p2pNode's public key hash
	nodePubKey := eth_crypto.FromECDSAPub(&prvEcdsaKey.PublicKey)
	pubKeyHash := eth_crypto.Keccak256Hash(nodePubKey[1:]).String()

	// Create a new p2pNode config
	conf, err := p2pCfg.ToNodeConfig()
	if err != nil {
		log.Panicf("invalid p2p config: %v", err)
	}

	// Create a new p2pNode instance
	nodeHost, err := node.NewNode(ctx, log, conf)
	if err != nil {
		log.Panicf("cannot initiate p2p p2pNode: %v", err)
	}
	p2pNode := &Node{
		ctx:  ctx,
		Node: nodeHost,
		log:  log,

		prvSecp256k1Key: prvSecp256k1Key.(*crypto.Secp256k1PrivateKey),
		prvEcdsaKey:     prvEcdsaKey,
		pubKeyHash:      pubKeyHash,

		emitBeaconMessages: true,
		EnableSync:         p2pCfg.EnableSync,

		obsReceiveRes: obsReceiveRes,
		obsSendReq:    obsSendReq,
	}

	log.WithFields(logrus.Fields{
		"id":    p2pNode.Node.ID(),
		"addrs": p2pNode.Node.ListenAddrs(),
	}).Info("P2P node started")

	return p2pNode
}

// GetPeerID returns the node's peer ID.
func (n *Node) GetPeerID() peer.ID {
	return n.Node.ID()
}

func (n *Node) GetAddress() string {
	return n.pubKeyHash[26:]
}

// HasPeers returns true if the node has at least one peer.
func (n *Node) HasPeers() bool {
	if n.Node.Peerstore().Peers().Len() > 0 {
		return true
	}

	n.log.Info("Only one peer available")
	return false
}

// IsPeerConnected returns true if the peer is connected to the node.
func (n *Node) IsPeerConnected(peerID peer.ID) bool {
	return n.Node.Network().Connectedness(peerID) == p2pNetwork.Connected
}

func (n *Node) Start(wg *sync.WaitGroup) {

	relayerHandlers := map[string]MessageHandler{}

	n.log.Infof("Enabling go-scammer topic")
	relayerHandlers["go-scammer"] = &NewMessageHandler{
		n: n,
	}

	n.Node.ConnectToBootstrapPeers()
	// Only the main topic is necessary for setting up peer discovery, since every relayer connects to it
	n.Node.SetUpPeerDiscovery(RelayersPubsubTopic)

	for topic := range relayerHandlers {
		if err := n.Node.JoinTopic(topic); err != nil {
			n.log.WithError(err).Panicf("cannot join %s topic", topic)
		}
	}

	// Start node event loop after joining the topics
	n.Node.Start(wg)

	// Relayers message loop
	for topic := range relayerHandlers {
		wg.Add(1)
		go relayerHandlers[topic].MessageLoop(topic, wg)
	}
	wg.Add(1)
	go n.PublishMessageLoop("go-scammer", wg)
}
