package handlers

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-lib/mqtt_manager"
	"github.com/rs/zerolog/log"
)

type deviceOnlineHandler struct {
	device        IDeviceService
	notifications INotificationService
}

func DeviceOnline(device IDeviceService, notifications INotificationService) IHandlers {
	return &deviceOnlineHandler{device: device, notifications: notifications}
}

// Handle processes device status change messages
//
// topic: device/{DEVICE-UID}/online
func (h *deviceOnlineHandler) Handle(_ mqtt.Client, msg mqtt_manager.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Debug().Msg("extract message params")
	uid, err := msg.GetUid()
	if err != nil {
		log.Error().Err(err).Msg("failed to get message uid")
		return
	}

	log.Debug().Str("uid", uid).Msg("get data from db")
	device, err := h.device.GetDevice(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting device")
		return
	}

	chatIds, err := h.device.GetDeviceChats(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting chats")
		return
	}

	log.Debug().Str("uid", uid).Msg("sending status message")
	args := map[string]interface{}{
		"Label":     device.Label,
		"Point":     device.Point,
		"UID":       device.UID,
		"ChangedAt": device.LastSeen,
	}

	errs := h.notifications.SendToChats("device-online.tmpl", chatIds, args)
	for chatId, e := range errs {
		if e != nil {
			log.Warn().Int64("chat", chatId).Err(e).Msg("error sending status notification")
		}
	}
}
