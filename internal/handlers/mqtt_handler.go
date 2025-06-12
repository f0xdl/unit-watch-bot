package handlers

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"time"
)

type MqttHandler struct {
	bot       *tgbotapi.BotAPI
	templater *templates.TemplateService
	chatId    int64
}

func NewMqttHandler(bot *tgbotapi.BotAPI, templater *templates.TemplateService, chatId int64) *MqttHandler {
	return &MqttHandler{bot, templater, chatId}
}

func (mh *MqttHandler) statusHandler(_ mqtt.Client, msg mqtt.Message) {
	args := map[string]string{
		"DeviceName": "aaaa",
		"Event":      "status changed to " + string(msg.Payload()),
		"Time":       time.Now().Format(time.RFC3339),
	}

	printableMsg, err := mh.templater.Render("device_alert.tmpl", args)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering template")
		return
	}
	msgTme := tgbotapi.NewMessage(mh.chatId, printableMsg)
	result, err := mh.bot.Send(msgTme)
	if err != nil {
		log.Error().Err(err).Msg("Error sending message")
		return
	}
	log.Info().Any("msg", result).Msg("Message sent")
}

func (mh *MqttHandler) SubscribeTopics(client mqtt.Client) {
	topics := map[string]mqtt.MessageHandler{
		"device/+/status": mh.statusHandler,
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
