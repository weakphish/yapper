package db

import (
	"fmt"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"github.com/weakphish/yapper/internal/config"
	"github.com/weakphish/yapper/internal/model"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	dbPath := config.YapConfig.GetString(config.DbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	db.AutoMigrate(&model.Task{})

	return db, nil
}
