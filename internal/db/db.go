package db

import (
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)


func ProvideDB(logger zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		logger.Error("Database could not be opened.", zap.Error(err))
		return nil, err
	}

	// db.AutoMigrate();

	return db, nil
}