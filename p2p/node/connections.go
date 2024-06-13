package node

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/sirupsen/logrus"
)

const reconnectInterval = 60 * time.Second
const connectionTimeout = 5 * time.Second

type ConnectionManager struct {
	ctx context.Context
	log *logrus.Entry

	host host.Host
}

func NewConnectionManager(ctx context.Context, log *logrus.Entry, host host.Host) *ConnectionManager {
	return &ConnectionManager{
		ctx:  ctx,
		log:  log,
		host: host,
	}
}

// connectToPeers connects to all peers in the peerstore (if not already connected)
func (cm *ConnectionManager) connectToPeers() {
	var wg sync.WaitGroup
	for _, peerID := range cm.host.Peerstore().Peers() {
		// Don't connect to self
		if peerID == cm.host.ID() {
			continue
		}

		// Don't connect to peers we're already connected to
		if cm.host.Network().Connectedness(peerID) != network.NotConnected {
			continue
		}

		// Connect/reconnect to peer
		wg.Add(1)
		peer := cm.host.Peerstore().PeerInfo(peerID)
		logPeer := cm.log.WithField("peer", peer)

		logPeer.Debug("Re-connecting to node")
		go func() {
			defer wg.Done()

			timeoutCtx, cancel := context.WithTimeout(cm.ctx, connectionTimeout)
			defer cancel()

			if err := cm.host.Connect(timeoutCtx, peer); err != nil {
				logPeer.Warnf("Error while connecting to peer: %-v", err)
				cm.host.Peerstore().RemovePeer(peer.ID)
				cm.host.Peerstore().ClearAddrs(peer.ID)
				return
			}

			logPeer.Info("Connection established with peer")
		}()
	}
	wg.Wait()
}

func (cm *ConnectionManager) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(reconnectInterval)
	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.connectToPeers()
		}
	}
}
