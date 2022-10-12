package data

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const databaseName = "lfg"

// CreateConnection creates the database connection object.
func CreateConnection() (*gorm.DB, error) {
	databaseFile := fmt.Sprintf("./%s.db", databaseName)
	db, err := gorm.Open(sqlite.Open(databaseFile), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not open SQL database. Error: %v", err)
		return nil, err
	}
	log.Info("Created database connection sucessfully")
	return db, nil
}
