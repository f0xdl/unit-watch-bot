package storage

import (
	"gorm.io/gorm"
	"time"
)

type Device struct {
	gorm.Model
	UUID      string   `gorm:"uniqueIndex"`
	Active    bool     `gorm:"default:false;not null;"`
	Status    int      `gorm:"default:0;index;not null;"`
	Groups    []*Group `gorm:"many2many:device_groups;"`
	OwnerId   int64
	LastSeen  time.Time
	ExpiresAt time.Time
}

type Group struct {
	gorm.Model
	ChatID  int64     `gorm:"uniqueIndex"`
	Devices []*Device `gorm:"many2many:device_groups;"`
}

type Owner struct {
	gorm.Model
	UserId  int64     `gorm:"uniqueIndex"`
	Devices []*Device `gorm:"foreignkey:OwnerId;references:ID"`
}
