package services

import (
	"context"
	"github.com/f0xdl/unit-watch-lib/domain"
	"github.com/f0xdl/unit-watch-lib/repository"
)

type DeviceService struct {
	deviceStore *repository.DeviceRepository
}

func NewDeviceService(deviceStore *repository.DeviceRepository) *DeviceService {
	return &DeviceService{
		deviceStore: deviceStore,
	}
}

func (ds *DeviceService) GetDevice(ctx context.Context, uid string) (domain.Device, error) {
	return ds.deviceStore.Get(ctx, uid)
}

func (ds *DeviceService) HasStatusChanged(device domain.Device, newStatus domain.DeviceStatus) bool {
	return device.Status != newStatus
}

func (ds *DeviceService) GetDeviceChatIds(ctx context.Context, uid string) ([]int64, error) {
	return ds.deviceStore.GetChatIds(ctx, uid)
}
func (ds *DeviceService) GetDeviceChats(ctx context.Context, uid string) (map[int64]string, error) {
	return ds.deviceStore.GetDeviceChats(ctx, uid)
}

func (ds *DeviceService) UpdateDeviceStatus(ctx context.Context, uid string, status domain.DeviceStatus) error {
	return ds.deviceStore.UpdateStatus(ctx, uid, status)
}
