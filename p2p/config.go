package p2p

import (
	"fmt"

	"gossip-protocol/config"
	"gossip-protocol/p2p/node"
)

type Config struct {
	ListenAddresses        []string
	BootstrapNodeAddresses []string
	PrivateKey             string
	DiscoveryMethods       []string
	EnableSync             bool
}

func ToP2pConfig(config config.P2PConfig) Config {
	return Config{
		PrivateKey:             config.PrivateKey,
		ListenAddresses:        config.ListenAddresses,
		BootstrapNodeAddresses: config.BootstrapNodeAddresses,
		DiscoveryMethods:       config.DiscoveryMethods,
		EnableSync:             config.EnableSync,
	}
}

func (c Config) ToNodeConfig() (node.Config, error) {
	privateKey, err := node.ToLibp2pPrivateKey(c.PrivateKey)
	if err != nil {
		return node.Config{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	listenAddrs, err := node.ToMultiaddrs(c.ListenAddresses)
	if err != nil {
		return node.Config{}, fmt.Errorf("failed to parse listen addresses: %w", err)
	}

	discoveryMethods, err := node.ToDiscoveryMethod(c.DiscoveryMethods)
	if err != nil {
		return node.Config{}, fmt.Errorf("failed to parse discovery methods: %w", err)
	}

	bootstrapNodes, err := node.ToAddrInfo(c.BootstrapNodeAddresses)
	if err != nil {
		return node.Config{}, fmt.Errorf("failed to parse bootstrap node addresses: %w", err)
	}

	conf := node.Config{
		PrivateKey:       privateKey,
		ListenAddresses:  listenAddrs,
		DiscoveryMethods: discoveryMethods,
		BootstrapPeers:   bootstrapNodes,
	}
	return conf, nil
}
