package handlers

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	"github.com/f0xdl/unit-watch-lib/domain"
	"github.com/f0xdl/unit-watch-lib/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

type MqttHandler struct {
	bot         *tgbotapi.BotAPI
	templater   *templates.TemplateService
	deviceStore *storage.GormStorage
}

func NewMqttHandler(bot *tgbotapi.BotAPI, templater *templates.TemplateService, deviceStore *storage.GormStorage) *MqttHandler {
	return &MqttHandler{bot: bot, templater: templater, deviceStore: deviceStore}
}

func (mh *MqttHandler) SubscribeTopics(client mqtt.Client) {
	topics := map[string]mqtt.MessageHandler{
		"device/+/status": mh.StatusHandler,
		"device/+/online": mh.OnlineHandler,
	}
	for topic, handler := range topics {
		token := client.Subscribe(topic, 1, handler)
		if token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Str("topic", topic).Msg("error subscribing to topic")
		} else {
			log.Info().Str("topic", topic).Msg("subscribed to topic")
		}
	}
}

// StatusHandler topic: device/{DEVICE-UID}/status
func (mh *MqttHandler) StatusHandler(_ mqtt.Client, msg mqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uid := getUID(msg)
	sInt, err := strconv.Atoi(string(msg.Payload()))
	if err != nil {
		log.Error().Msgf("error converting status to int: %s", string(msg.Payload()))
	}
	status := domain.ParseDeviceStatus(sInt)

	device, err := mh.deviceStore.Get(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting device")
		return
	}
	if device.Status == status {
		log.Warn().
			Str("uid", uid).
			Any("status", device.Status).
			Msg("device status not changed")
		return
	}

	chatIds, err := mh.deviceStore.GetChatIds(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting chats")
		return
	}

	args := map[string]interface{}{
		"Online":      device.Online,
		"Label":       device.Label,
		"Point":       device.Point,
		"UID":         device.UID,
		"OldStatus":   device.Status,
		"NewStatus":   status,
		"ChangedAt":   templates.FormatChangedAt(time.Now(), "uk"),
		"LastSeenAgo": templates.FormatSeenAgo(device.LastSeen, "uk"),
	}
	for k, v := range args {
		switch v.(type) {
		case string:
			args[k] = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, v.(string))
		}
	}

	printableMsg, err := mh.templater.Render("device-new-status-notify.tmpl", args)
	if err != nil {
		log.Error().Err(err).Msg("error rendering template")
		return
	}

	for _, chatId := range chatIds {
		msgTme := tgbotapi.NewMessage(chatId, printableMsg)
		msgTme.ParseMode = tgbotapi.ModeMarkdownV2
		_, err = mh.bot.Send(msgTme)
		if err != nil {
			log.Error().Err(err).Int64("chatId", chatId).Msg("error sending message")
			continue
		}
		log.Info().Int64("chatId", chatId).Msg("message sent")
	}
	err = mh.deviceStore.UpdateStatus(ctx, uid, status)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error updating status")
		return
	}
}

// OnlineHandler topic: device/{DEVICE-UID}/online
func (mh *MqttHandler) OnlineHandler(_ mqtt.Client, msg mqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	uid := getUID(msg)
	online, err := strconv.ParseBool(string(msg.Payload()))
	if err != nil {
		log.Error().Msgf("error converting status to int: %s", string(msg.Payload()))
	}

	device, err := mh.deviceStore.Get(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting device")
		return
	}

	if device.Online == online {
		log.Warn().
			Str("uid", uid).
			Any("status", device.Status).
			Msg("device already online")
		return
	}

	chatIds, err := mh.deviceStore.GetChatIds(ctx, uid)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error getting chats")
		return
	}
	args := map[string]interface{}{
		"Online": online,
		"Label":  device.Label,
		"Point":  device.Point,
		"UID":    device.UID,
	}
	for k, v := range args {
		switch v.(type) {
		case string:
			args[k] = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, v.(string))
		}
	}
	printableMsg, err := mh.templater.Render("device-online.tmpl", args)
	if err != nil {
		log.Error().Err(err).Msg("error rendering template")
		return
	}
	for _, chatId := range chatIds {
		msgTme := tgbotapi.NewMessage(chatId, printableMsg)
		msgTme.ParseMode = tgbotapi.ModeMarkdownV2
		_, err = mh.bot.Send(msgTme)
		if err != nil {
			log.Error().Err(err).Int64("chatId", chatId).Msg("error sending message")
			continue
		}
		log.Info().Int64("chatId", chatId).Msg("message sent")
	}

	err = mh.deviceStore.UpdateOnline(ctx, uid, online)
	if err != nil {
		log.Error().Err(err).Str("uid", uid).Msg("error updating status")
		return
	}
}

func getUID(msg mqtt.Message) string {
	topicParts := strings.Split(msg.Topic(), "/")
	if len(topicParts) != 3 {
		log.Error().Msgf("topic format error: %s", msg.Topic())
	}
	return topicParts[1]
}
