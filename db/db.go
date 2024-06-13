package db

import (
	"gossip-protocol/db/models"
	"gossip-protocol/network"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBservices interface {
	Add(msg *network.Message) error
	Delete(id string) error
	Message(id string) (*network.Message, error)
	GetAll() []*network.Message
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

func (db *PostgresDbService) Add(msg *models.DBMessages, maxMessages int64) error {

	if result := db.Db.Create(msg); result.Error != nil {
		return result.Error
	}
	return nil
}

func (db *PostgresDbService) Delete(id string) error {
	if result := db.Db.Where("id = ?", id).Delete(&models.DBMessages{}); result.Error != nil {
		return result.Error
	}
	return nil
}

func (db *PostgresDbService) GetMessageById(id string) (*models.DBMessages, error) {
	var msg models.DBMessages
	if result := db.Db.Where("id = ?", id).Limit(1).Take(&msg); result.Error != nil {
		return &models.DBMessages{}, result.Error
	}
	return &msg, nil
}

func (db *PostgresDbService) GetAllMessages() ([]*models.DBMessages, error) {
	messages := make([]*models.DBMessages, 0)

	if result := db.Db.Model(&models.DBMessages{}).Find(&messages); result.Error != nil {
		return messages, result.Error
	}
	return messages, nil
}
