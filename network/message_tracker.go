package network

import (
	"errors"
	database "gossip-protocol/db"
	"gossip-protocol/db/models"
	"time"
)

// MessageTracker tracks a configurable fixed amount of messages.
// Messages are stored first-in-first-out.  Duplicate messages should not be stored in the queue.
type MessageTracker interface {
	// Add will add a message to the tracker, deleting the oldest message if necessary
	Add(message *Message) (err error)
	// Delete will delete message from tracker
	Delete(id string) (err error)
	// Get returns a message for a given ID.  Message is retained in tracker
	Message(id string) (message *Message, err error)
	// Messages returns messages in FIFO order
	Messages() (messages []*Message)
}

// ErrMessageNotFound is an error returned by MessageTracker when a message with specified id is not found
var ErrMessageNotFound = errors.New("message not found")

func NewMessageTracker(length int, db database.DBservices) MessageTracker {
	// TODO: Implement this constructor with your implementation of the MessageTracker interface

	tracker := newTracker(length, db)

	return tracker
}

type Tracker struct {
	db          database.DBservices
	maxMessages int
}

func newTracker(maxMessages int, db database.DBservices) *Tracker {
	return &Tracker{
		db:          db,
		maxMessages: maxMessages,
	}
}
func (t *Tracker) Add(message *Message) (err error) {
	newDbMessage := &models.DBMessages{
		MsgID:     message.ID,
		PeerID:    message.PeerID,
		Data:      message.Data,
		CreatedAt: time.Now().Unix(),
	}
	if err = t.db.Add(newDbMessage, t.maxMessages); err != nil {
		return err
	}

	return nil
}

func (t *Tracker) Delete(id string) (err error) {
	if err := t.db.Delete(id); err != nil {
		if err.Error() == "record not found" {
			return ErrMessageNotFound
		} else {
			return err
		}
	}
	return nil
}

func (t *Tracker) Message(id string) (message *Message, err error) {
	dbMessage, err := t.db.GetMessageById(id)
	message = &Message{}
	if err != nil {
		if err.Error() == "record not found" {
			return nil, ErrMessageNotFound
		} else {
			return nil, err
		}
	}
	message.ID = dbMessage.MsgID
	message.PeerID = dbMessage.PeerID
	message.Data = dbMessage.Data
	return message, nil
}

func (t *Tracker) Messages() (messages []*Message) {
	dbMessages := t.db.GetAll()
	messages = make([]*Message, 0)
	for _, dbMessage := range dbMessages {
		newMessage := &Message{
			ID:     dbMessage.MsgID,
			Data:   dbMessage.Data,
			PeerID: dbMessage.PeerID,
		}
		messages = append(messages, newMessage)

	}

	return messages
}
