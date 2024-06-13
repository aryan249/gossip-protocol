package models

type DBMessages struct {
	ID        string `json:"id"`
	PeerID    string `json:"peer_id"`
	Data      []byte `json:"data"`
	CreatedAt int64  `json:"created_at"`
}
