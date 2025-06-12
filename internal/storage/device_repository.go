package storage

import (
	"context"
	"github.com/f0xdl/unit-watch-bot/internal/domain"
	"gorm.io/gorm"
	"time"
)

type GormStorage struct {
	db *gorm.DB
}

func NewGormStorage(db *gorm.DB) *GormStorage {
	return &GormStorage{db: db}
}

func (s *GormStorage) GetStatus(ctx context.Context, uuid string) (domain.DeviceStatus, error) {
	var d Device
	err := s.db.
		WithContext(ctx).
		First(&d, "uuid = ?", uuid).
		Error
	if err != nil {
		return 0, err
	}
	return domain.DeviceStatus(d.Status), nil
}

func (s *GormStorage) UpdateStatus(ctx context.Context, uuid string, status int) error {
	return s.db.WithContext(ctx).
		Model(&Device{}).
		Where("uuid = ?", uuid).
		Update("status", status).Error
}

func (s *GormStorage) UpdateOnline(ctx context.Context, uuid string, at time.Time) error {
	return s.db.WithContext(ctx).
		Model(&Device{}).
		Where("uuid = ?", uuid).
		Update("last_seen", at).
		Error
}

func (s *GormStorage) Get(ctx context.Context, uuid string) (domain.Device, error) {
	var d Device
	err := s.db.WithContext(ctx).
		First(&d, "uuid = ?", uuid).
		Error
	if err != nil {
		return domain.Device{}, err
	}
	return domain.Device{
		UUID:     d.UUID,
		Status:   domain.DeviceStatus(d.Status),
		LastSeen: d.LastSeen,
		Active:   d.Active,
	}, nil
}

func (s *GormStorage) GetChatIds(ctx context.Context, uuid string) ([]int64, error) {
	var d Device
	err := s.db.WithContext(ctx).
		Preload("Groups").
		First(&d, "uuid = ?", uuid).
		Error
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(d.Groups))
	for i, g := range d.Groups {
		ids[i] = g.ChatID
	}
	return ids, nil
}
