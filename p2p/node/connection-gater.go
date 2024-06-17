// Implementation of the ConnectionGater interface that will prevent any inbound node connection from a node that does not belong to the whitelist.
// The whitelist is based on a list of authorized PerpApi addresses

package node

import (
	"context"
	"strings"

	// "github.com/ethereum/go-ethereum/crypto"
	ma "github.com/multiformats/go-multiaddr"

	// "crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
)

type ConnectionGater struct {
	ctx            context.Context
	log            *logrus.Entry
	TrustedAddress []common.Address
}

func NewConnectionGater(ctx context.Context, log *logrus.Entry) connmgr.ConnectionGater {
	return &ConnectionGater{
		ctx: ctx,
		log: log,
	}
}

// InterceptPeerDial is called on an imminent outbound peer dial request, prior
// to the addresses of that peer being available/resolved. Blocking connections
// at this stage is typical for blacklisting scenarios.
func (c ConnectionGater) InterceptPeerDial(p peer.ID) (allow bool) {
	// We always allow outbound requests
	return true
}

// InterceptAddrDial is called on an imminent outbound dial to a peer on a
// particular address. Blocking connections at this stage is typical for
// address filtering.
func (c ConnectionGater) InterceptAddrDial(peer.ID, ma.Multiaddr) (allow bool) {
	// We always allow outbound requests
	return true
}

// InterceptAccept is called as soon as a transport listener receives an
// inbound connection request, before any upgrade takes place. Transports who
// accept already secure and/or multiplexed connections (e.g. possibly QUIC)
// MUST call this method regardless, for correctness/consistency.
func (c ConnectionGater) InterceptAccept(n network.ConnMultiaddrs) (allow bool) {
	return true
}

// InterceptSecured is called for both inbound and outbound connections,
// after a security handshake has taken place and we've authenticated the peer.
func (c ConnectionGater) InterceptSecured(n network.Direction, peer peer.ID, nn network.ConnMultiaddrs) (allow bool) {
	return true
}

// InterceptUpgraded is called for inbound and outbound connections, after
// libp2p has finished upgrading the connection entirely to a secure,
// multiplexed channel.
func (c ConnectionGater) InterceptUpgraded(n network.Conn) (allow bool, reason control.DisconnectReason) {

	// determine public key of peer

	pub, err := n.RemotePublicKey().Raw()

	if err != nil {
		c.log.Error(err)
		return false, 0
	}

	publicKeyECDSA, err := crypto.DecompressPubkey(pub)

	if err != nil {
		c.log.Error(err)
		return false, 0
	}

	// Extract address associated with the public key
	fromAddress := strings.ToLower(crypto.PubkeyToAddress(*publicKeyECDSA).String())

	isAuthorized := false

	if len(c.TrustedAddress) > 0 {
		for index := 0; index < len(c.TrustedAddress); index++ {
			if strings.ToLower(c.TrustedAddress[index].Hex()) == fromAddress {
				c.log.Debugf("Authorizing peer %s based on Perp API EOA address", n.RemotePeer())
				isAuthorized = true
				break
			}
		}
	}

	c.log.Infof("Peer %s (address %s) authorized? %t", n.RemotePeer(), fromAddress, isAuthorized)
	// We always give the same reason 0. This is because this is an experimental type and setting it
	// to any particular value other than 0 is moot at the moment
	return isAuthorized, 0
}
