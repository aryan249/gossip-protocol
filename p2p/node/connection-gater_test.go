package node

import (
	"context"
	"fmt"
	"testing"
	"time"

	mock "github.com/libp2p/go-libp2p/p2p/net/mock"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type MockNetworkConn struct {
	ctx context.Context
	log *logrus.Entry

	guardianSetMapping *map[string][]GuardianInfo
}

func TestConnectionGater(t *testing.T) {

	logger := NewMockLogger("debug")

	ctx := context.Background()

	guardianSetMapping := make(map[string][]GuardianInfo)

	var guardianSet []GuardianInfo
	guardianSet = append(guardianSet, GuardianInfo{Address: string("0x74bC67ed6948f0a4C387C353975F142Dc640537a")})
	guardianSet = append(guardianSet, GuardianInfo{Address: string("0x7c610B4dDA11820b749AeA40Df8cBfdA1925e581")})
	guardianSet = append(guardianSet, GuardianInfo{Address: string("0xC856bCA66CEE14FdfaF6Bb73c3f3bfd4bEfB61E5")})
	guardianSet = append(guardianSet, GuardianInfo{Address: string("0xa2dAC41C6b3e924524aB6E00f775E8Dd6f7f506A")})
	guardianSet = append(guardianSet, GuardianInfo{Address: string("0xa970dc68F993bb3Bda4b36685ed6BA44d9C997D1")})

	guardianSetMapping["ARB"] = guardianSet

	connGater := NewConnectionGater(ctx, logger.WithField("service", "connection_gater"))

	peer1PrivateKey, _ := ToLibp2pPrivateKey("bb39159966fde2b2353316e2ea0f58da252db67cec22430c6a18f48787d90d02")
	peer2PrivateKey, _ := ToLibp2pPrivateKey("213f1d9567f011e10c4143d892ac799595c6f16fa3f044c8278a304dbd12ddc4")

	mockNet := mock.New()

	peer1Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50001))
	peer2Address, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", 50002))

	peer1Host, _ := mockNet.AddPeer(peer1PrivateKey, peer1Address)
	peer2Host, _ := mockNet.AddPeer(peer2PrivateKey, peer2Address)

	peer1Network := peer1Host.Network()
	peer2Network := peer2Host.Network()

	peer1 := peer1Host.ID()
	peer2 := peer2Host.ID()

	mockNet.LinkPeers(peer1, peer2)

	n12, _ := peer1Network.DialPeer(ctx, peer2)
	n21, _ := peer2Network.DialPeer(ctx, peer1)

	isAuthorized12, _ := connGater.InterceptUpgraded(n12)
	isAuthorized21, _ := connGater.InterceptUpgraded(n21)
	assert.False(t, isAuthorized12, "Connection should not be authorized if the address is not contained in the list of guardian owners")
	assert.False(t, isAuthorized21, "Connection should not be authorized if the address is not contained in the list of guardian owners")

	// Adding the address of peer2 as the public address of perpApi should allow connection 1 -> 2.
	// This is because the connection we are testing here is outbound, and from the point of view of peer 1. For an inbound connection 1 -> 2,
	// the connection gater would check peer 1's public key, from peer 2's point of view

	connGater = NewConnectionGater(ctx, logger.WithField("service", "connection_gater"))
	isAuthorized21, _ = connGater.InterceptUpgraded(n21)
	assert.False(t, isAuthorized21, "Connection 2 -> 1 should not be authorized, as only peer2 is in the list of PerpApi addresses")
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
