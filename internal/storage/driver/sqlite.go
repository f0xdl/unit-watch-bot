package driver

import (
	"github.com/f0xdl/unit-watch-bot/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitSQLite(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&storage.Device{}, &storage.Group{}, &storage.Owner{}); err != nil {
		return nil, err
	}
	return db, nil
}
