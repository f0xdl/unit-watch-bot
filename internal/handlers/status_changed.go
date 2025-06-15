package handlers

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-lib/mqtt_manager"
	"github.com/rs/zerolog/log"
	"time"
)

type statusChangedHandler struct {
	device        IDeviceService
	notifications INotificationService
}

func StatusChanged(device IDeviceService, notifications INotificationService) IHandlers {
	return &statusChangedHandler{device: device, notifications: notifications}
}

// Handle processes device status change messages
//
// topic: device/{DEVICE-UID}/status
func (h *statusChangedHandler) Handle(_ mqtt.Client, msg mqtt_manager.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Debug().Msg("extract message params")
	uid, err := msg.GetUid()
	if err != nil {
		log.Error().Err(err).Msg("failed to get message uid")
		return
	}
	newStatus, err := msg.GetDeviceStatus()
	if err != nil {
		log.Error().Err(err).Msg("failed to get message status")
		return
	}
	log.Debug().Str("uid", uid).Msg("get data from db")
	device, err := h.device.GetDevice(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting device")
		return
	}

	if !h.device.HasStatusChanged(device, newStatus) {
		log.Warn().
			Str("uid", uid).
			Any("current_status", device.Status).
			Any("new_status", newStatus).
			Msg("device status not changed")
		return
	}

	chatIds, err := h.device.GetDeviceChats(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting chats")
		return
	}

	log.Debug().Str("uid", uid).Msg("sending status message")
	args := map[string]interface{}{
		"Online":    device.Online,
		"Label":     device.Label,
		"Point":     device.Point,
		"UID":       device.UID,
		"OldStatus": device.Status,
		"NewStatus": newStatus,
		"ChangedAt": time.Now(),
		"LastSeen":  device.LastSeen,
	}
	errs := h.notifications.SendToChats("device-status-changed.tmpl", chatIds, args)
	for chatId, e := range errs {
		if e != nil {
			log.Warn().Int64("chat", chatId).Err(e).Msg("error sending status notification")
		}
	}
	if len(errs) == len(chatIds) { //TODO: doc this block
		log.Error().Msg("error sending status notification")
		return
	}

	log.Debug().Str("uid", uid).Msg("update device status")
	err = h.device.UpdateDeviceStatus(ctx, uid, newStatus)
	if err != nil {
		log.Error().Err(err).Msg("error updating device status")
	}
	log.Debug().
		Str("uid", uid).
		Any("old_status", device.Status).
		Any("new_status", newStatus).
		Msg("device status updated successfully")
}
