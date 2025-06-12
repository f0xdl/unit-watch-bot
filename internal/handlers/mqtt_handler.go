package handlers

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-bot/internal/domain"
	"github.com/f0xdl/unit-watch-bot/internal/storage"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
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

func getUUID(msg mqtt.Message) string {
	topicParts := strings.Split(msg.Topic(), "/")
	if len(topicParts) != 3 {
		log.Error().Msgf("Topic format error: %s", msg.Topic())
	}
	return topicParts[1]
}

func (mh *MqttHandler) StatusHandler(_ mqtt.Client, msg mqtt.Message) {
	// topic: device/{DEVICE-UUID}/status
	uuid := getUUID(msg)
	sInt, err := strconv.Atoi(string(msg.Payload()))
	if err != nil {
		log.Error().Msgf("Error converting status to int: %s", string(msg.Payload()))
	}
	status := domain.ParseDeviceStatus(sInt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	device, err := mh.deviceStore.Get(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Error getting device")
		return
	}
	chatIds, err := mh.deviceStore.GetChatIds(ctx, uuid)
	if err != nil {
		log.Error().Err(err).Str("uuid", uuid).Msg("Error getting device")
		return
	}

	args := map[string]interface{}{
		"Online":      device.Online(),
		"Label":       device.Label,
		"Point":       device.PointId,
		"UUID":        device.UUID,
		"OldStatus":   device.Status,
		"NewStatus":   status,
		"ChangedAt":   time.Now().Format(time.RFC3339),
		"LastSeenAgo": device.LastSeen,
	}

	printableMsg, err := mh.templater.Render("device-status-notify.tmpl", args)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering template")
		return
	}

	for _, chatId := range chatIds {
		msgTme := tgbotapi.NewMessage(chatId, printableMsg)
		_, err = mh.bot.Send(msgTme)
		if err != nil {
			log.Error().Err(err).Int64("chatId", chatId).Msg("Error sending message")
			continue
		}
		log.Info().Int64("chatId", chatId).Msg("Message sent")
	}
}

func (mh *MqttHandler) SubscribeTopics(client mqtt.Client) {
	topics := map[string]mqtt.MessageHandler{
		"device/+/status": mh.StatusHandler,
	}
	for topic, handler := range topics {
		token := client.Subscribe(topic, 1, handler)
		if token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Str("topic", topic).Msg("Error subscribing to topic")
		} else {
			log.Info().Str("topic", topic).Msg("Subscribed to topic")
		}
	}
}
