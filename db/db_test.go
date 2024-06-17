package db

import (
	"gossip-protocol/config"
	"gossip-protocol/db/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageAddition(t *testing.T) {
	cfg := config.NewViperConfig()
	dbConfig := cfg.ReadDBConfig()
	url := dbConfig.AsPostgresDbUrl()
	db := Init(url)
	maxlength := 5
	messages := make([]*models.DBMessages, 0)
	db.Db.Model(&models.DBMessages{}).Find(&messages)

	db.Db.Delete(&messages)

	msg := &models.DBMessages{
		MsgID:  "TestMsgId",
		PeerID: "TestPerrId123",
		Data:   []byte{},
	}

	db.Add(msg, maxlength)

	result, err := db.GetMessageById(msg.MsgID)
	assert.NoError(t, err)
	assert.Equal(t, result.PeerID, msg.PeerID)

	// If messaged is not prevert it will revert
	_, err = db.GetMessageById("29049202020")

	assert.NotEmpty(t, err)
}

func TestDeleteMessage(t *testing.T) {
	cfg := config.NewViperConfig()
	dbConfig := cfg.ReadDBConfig()
	url := dbConfig.AsPostgresDbUrl()
	db := Init(url)
	maxlength := 5
	messages := make([]*models.DBMessages, 0)
	db.Db.Model(&models.DBMessages{}).Find(&messages)

	db.Db.Delete(&messages)

	msg := &models.DBMessages{
		MsgID:  "TestMsgId",
		PeerID: "TestPerrId123",
		Data:   []byte{},
	}

	db.Add(msg, maxlength)

	err := db.Delete(msg.MsgID)
	assert.NoError(t, err)

	// If messaged is not prevert it will revert
	_, err = db.GetMessageById("29049202020")

	assert.NotEmpty(t, err)
}

func TestGetAllMessages(t *testing.T) {
	cfg := config.NewViperConfig()
	dbConfig := cfg.ReadDBConfig()
	url := dbConfig.AsPostgresDbUrl()
	db := Init(url)
	maxlength := 5
	messages := make([]*models.DBMessages, 0)
	db.Db.Model(&models.DBMessages{}).Find(&messages)

	db.Db.Delete(&messages)

	msg1 := &models.DBMessages{
		MsgID:  "TestMsgId",
		PeerID: "TestPerrId123",
		Data:   []byte{},
	}

	msg2 := &models.DBMessages{
		MsgID:  "TestMsgId2",
		PeerID: "TestPerrId1234",
		Data:   []byte{},
	}

	db.Add(msg1, maxlength)
	db.Add(msg2, maxlength)

	allMessages := db.GetAll()

	assert.Equal(t, len(allMessages), 2)

	assert.Equal(t, allMessages[0].MsgID, msg1.MsgID)
	assert.Equal(t, allMessages[1].MsgID, msg2.MsgID)

}
