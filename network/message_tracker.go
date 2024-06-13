package network

import (
	"errors"

	"github.com/jinzhu/gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func NewMessageTracker(length int, dbUrl string) MessageTracker {
	// TODO: Implement this constructor with your implementation of the MessageTracker interface

	var tracker MessageTracker

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})

	return Tracker{db: db}
}

type Tracker struct {
	db *gorm.DB
}

func (t *Tracker) Add(msg *Message) (err error) {

	return err
}

func (t *Tracker) Delete(id string) (err error) {
	return err
}

func (t *Tracker) Message(id string) (message *Message, err error) {
	return message, nil
}

func (t *Tracker) Messages(id string) (messages []*Message, err error) {
	return messages, nil
}
