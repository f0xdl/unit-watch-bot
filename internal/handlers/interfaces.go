package handlers

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-lib/domain"
	"github.com/f0xdl/unit-watch-lib/mqtt_manager"
)

type IDeviceService interface {
	GetDevice(ctx context.Context, uid string) (domain.Device, error)
	HasStatusChanged(device domain.Device, status domain.DeviceStatus) bool
	UpdateDeviceStatus(ctx context.Context, uid string, status domain.DeviceStatus) error
	GetDeviceChats(ctx context.Context, uid string) (map[int64]string, error)
}

type INotificationService interface {
	SendToChats(templateName string, chatIds map[int64]string, args map[string]interface{}) map[int64]error
}

type IHandlers interface {
	Handle(_ mqtt.Client, msg mqtt_manager.Message)
}
