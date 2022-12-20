package db

import (
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

func ProvideDB(logger *zap.Logger) (*gorm.DB, error) {
	gormLogger := zapgorm2.New(logger)
	gormLogger.IgnoreRecordNotFoundError = true
	gormLogger.SetAsDefault()

	db, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		logger.Error("Database could not be opened.", zap.Error(err))
		return nil, err
	}

	return db, nil
}
