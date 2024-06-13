package node

import "github.com/libp2p/go-libp2p/core/peer"

type GuardianInfo struct {
	Address string  `json:"pubkey"`
	PeerId  peer.ID `json:"peer_id"`
}
