package db

import (
	"gossip-protocol/db/models"
	"log"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBservices interface {
	Add(msg *models.DBMessages, maxMessages int) error
	Delete(id string) error
	GetMessageById(id string) (*models.DBMessages, error)
	GetAll() []*models.DBMessages
}

type PostgresDbService struct {
	Db *gorm.DB
}

func Init(dbUrl string) *PostgresDbService {

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.DBMessages{})

	return &PostgresDbService{Db: db}
}

func (db *PostgresDbService) Add(msg *models.DBMessages, maxMessages int) error {
	messages := make([]*models.DBMessages, 0)
	var count int64
	db.Db.Model(&models.DBMessages{}).Distinct("msg_id").Count(&count)

	msg.ID = strconv.Itoa(int(time.Now().Nanosecond())) + msg.MsgID
	db.Db.Model(&models.DBMessages{}).Order("created_at").Find(&messages)
	if count >= int64(maxMessages) {

		if deleteQuery := db.Db.Delete(messages[0]); deleteQuery.Error != nil {
			return deleteQuery.Error
		}

	}
	if result := db.Db.Create(msg); result.Error != nil {
		return result.Error
	}
	return nil
}

func (db *PostgresDbService) Delete(id string) error {
	var msg models.DBMessages
	if result := db.Db.Where("msg_id = ?", id).Limit(1).Take(&msg); result.Error != nil {
		return result.Error
	}
	if result := db.Db.Delete(&msg); result.Error != nil {
		return result.Error
	}
	println("messages with id deleted", id)
	return nil
}

func (db *PostgresDbService) GetMessageById(id string) (*models.DBMessages, error) {
	var msg models.DBMessages
	if result := db.Db.Where("msg_id = ?", id).Limit(1).Take(&msg); result.Error != nil {
		return &models.DBMessages{}, result.Error
	}
	return &msg, nil
}

func (db *PostgresDbService) GetAll() []*models.DBMessages {
	messages := make([]*models.DBMessages, 0)
	distinctMessages := make([]*models.DBMessages, 0)
	mapOfMessages := make(map[string]*models.DBMessages)
	if result := db.Db.Model(&models.DBMessages{}).Order("created_at").Find(&messages); result.Error != nil {
		log.Printf("error while fetching all messages %s", result.Error)
		return messages
	}

	for _, msg := range messages {
		if mapOfMessages[msg.MsgID] == nil {
			mapOfMessages[msg.MsgID] = msg
			distinctMessages = append(distinctMessages, msg)
		}
	}

	return distinctMessages
}
