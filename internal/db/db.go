package db

import (
	"fmt"

	"github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"github.com/weakphish/yapper/internal/model"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	// TODO: configure database path
	db, err := gorm.Open(sqlite.Open("yap.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error initializing database: %w", err)
	}

	db.AutoMigrate(&model.Task{}, &model.Note{})

	return db, nil
}
