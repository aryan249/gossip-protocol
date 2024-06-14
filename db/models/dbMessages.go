package models

type DBMessages struct {
	ID        string `gorm:"primaryKey;uniqueIndex:p;"json:"id"`
	MsgID     string `json:"msg_id"`
	PeerID    string `json:"peer_id"`
	Data      []byte `json:"data"`
	CreatedAt int64  `json:"created_at"`
}
