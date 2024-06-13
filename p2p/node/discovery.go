package node

import (
	"context"
	"fmt"
	"strings"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peerstore"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/sirupsen/logrus"
)

type DiscoveryMethod uint32

const (
	// DiscoveryMethodNone is the no discovery method
	DiscoveryMethodNone DiscoveryMethod = iota
	// DiscoveryMethodMDNS is the mDNS discovery method
	DiscoveryMethodMDNS
	// DiscoveryMethodDHT is the DHT discovery method
	DiscoveryMethodDHT
)

const discoveryInterval = 60 * time.Second

// ParseDiscoveryMethod takes a string discovery method and returns the discovery method constant.
func ParseDiscoveryMethod(method string) (DiscoveryMethod, error) {
	switch strings.ToLower(method) {
	case "none":
		return DiscoveryMethodNone, nil
	case "mdns":
		return DiscoveryMethodMDNS, nil
	case "dht":
		return DiscoveryMethodDHT, nil
	}

	var m DiscoveryMethod
	return m, fmt.Errorf("not a valid discovery method: %q", method)
}

// String returns the string representation of the discovery method.
func (dm DiscoveryMethod) String() string {
	if b, err := dm.MarshalText(); err == nil {
		return string(b)
	} else {
		return "unknown"
	}
}

// MarshalText returns the text representation of the discovery method.
func (dm DiscoveryMethod) MarshalText() ([]byte, error) {
	switch dm {
	case DiscoveryMethodNone:
		return []byte("none"), nil
	case DiscoveryMethodMDNS:
		return []byte("mdns"), nil
	case DiscoveryMethodDHT:
		return []byte("dht"), nil
	}

	return nil, fmt.Errorf("not a valid discovery method %d", dm)
}

// Initiates peer discovery, using the Kademlia DHT implementation, on the given topic
func DiscoverDHT(ctx context.Context, topic string, dht *dht.IpfsDHT, host host.Host, logger *logrus.Logger) {

	routingDiscovery := drouting.NewRoutingDiscovery(dht)
	dutil.Advertise(ctx, routingDiscovery, topic)

	ticker := time.NewTicker(discoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peerChan, err := routingDiscovery.FindPeers(ctx, topic)
			for peer := range peerChan {

				if peer.ID == host.ID() {
					continue
				}
				host.Peerstore().AddAddrs(peer.ID, peer.Addrs, peerstore.PermanentAddrTTL)
			}

			if err != nil {
				// There is nothing to do at this point. This is probably a transient network issue which will recover
				// at the next iteration. We at least log about it in order to help diagnostics
				logger.Errorf("Error while trying to find peers: %s", err)
			}
		}
	}

}
