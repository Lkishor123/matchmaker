package database

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"matchmaker/internal/logging"
)

var DB *gorm.DB

// Init opens a GORM connection using POSTGRES_URL. On success it sets DB.
func Init() (*gorm.DB, error) {
	dsn := os.Getenv("POSTGRES_URL")
	if dsn == "" {
		logging.Log.Warn("POSTGRES_URL not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logging.Log.WithError(err).Error("failed to connect to postgres")
		return nil, err
	}
	DB = db
	return db, nil
}
