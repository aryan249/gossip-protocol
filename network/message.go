package network

// Message is received from peers in a p2p network.
type Message struct {
	ID     string `json:"id"`
	PeerID string `json:"peer_id"`
	Data   []byte `json:"data"`
}
