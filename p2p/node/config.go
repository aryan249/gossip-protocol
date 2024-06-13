package node

import (
	"encoding/hex"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// ToMultiaddrs converts a string array to a Multiaddr array
func ToMultiaddrs(addresses []string) ([]multiaddr.Multiaddr, error) {
	addrs := make([]multiaddr.Multiaddr, 0, len(addresses))
	for _, addr := range addresses {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse multiaddress")
		}

		addrs = append(addrs, ma)
	}

	return addrs, nil
}

// ToAddrInfo converts a string array to a peer.AddrInfo array
func ToAddrInfo(addresses []string) ([]*peer.AddrInfo, error) {
	addrs := make([]*peer.AddrInfo, 0, len(addresses))
	for _, addr := range addresses {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse multiaddress")
		}

		info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse peer address info")
		}

		addrs = append(addrs, info)
	}

	return addrs, nil
}

func ToLibp2pPrivateKey(key string) (crypto.PrivKey, error) {
	bytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	prvKey, err := crypto.UnmarshalSecp256k1PrivateKey(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal private key: %w", err)
	}

	return prvKey, nil
}

// ToDiscoveryMethod converts a string array to a DiscoveryMethod array
func ToDiscoveryMethod(methods []string) ([]DiscoveryMethod, error) {
	discoveryMethods := make([]DiscoveryMethod, 0, len(methods))
	for _, method := range methods {
		m, err := ParseDiscoveryMethod(method)
		if err != nil {
			return nil, err
		}

		if m == DiscoveryMethodNone {
			return []DiscoveryMethod{DiscoveryMethodNone}, nil
		}

		if slices.Contains(discoveryMethods, m) {
			continue
		}

		discoveryMethods = append(discoveryMethods, m)
	}

	if len(discoveryMethods) == 0 {
		return []DiscoveryMethod{DiscoveryMethodNone}, nil
	}

	return discoveryMethods, nil
}

type Config struct {
	UserAgent        string
	PrivateKey       crypto.PrivKey
	ListenAddresses  []multiaddr.Multiaddr
	BootstrapPeers   []*peer.AddrInfo
	DiscoveryMethods []DiscoveryMethod
}
