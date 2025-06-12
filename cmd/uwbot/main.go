package main

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-bot/internal/configs"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"os/signal"
	"syscall"
	"time"
)

//// FSM для управления состояниями
//var botFSM = fsm.NewFSM(
//	"initial",
//	fsm.Events{
//		{Name: "start", Src: []string{"initial"}, Dst: "waiting"},
//		{Name: "message_received", Src: []string{"waiting"}, Dst: "processing"},
//		{Name: "processed", Src: []string{"processing"}, Dst: "waiting"},
//	},
//	fsm.Callbacks{},
//)

type MqttHandlers struct {
	bot       *tgbotapi.BotAPI
	templater *templates.TemplateService
	chatId    int64
}

func NewMqttHandlers(bot *tgbotapi.BotAPI, templater *templates.TemplateService, chatId int64) *MqttHandlers {
	return &MqttHandlers{bot, templater, chatId}
}

func (mh *MqttHandlers) statusHandler(client mqtt.Client, msg mqtt.Message) {
	args := map[string]string{
		"DeviceName": "aaaa",
		"Event":      "status changed to " + string(msg.Payload()),
		"Time":       time.Now().String(),
	}

	printableMsg, err := mh.templater.Render("device_alert.tmpl", args)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering template")
		return
	}
	msgTme := tgbotapi.NewMessage(mh.chatId, printableMsg)
	result, err := mh.bot.Send(msgTme)
	if err != nil {
		log.Error().Err(err).Msg("Error sending message: %s")
		return
	}
	log.Info().Any("msg", result).Msg("Message sent")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var err error

	log.Info().Msg("load configuration")
	cfg, err := configs.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config")
		return
	}

	//TODO: db
	//log.Info().Msg("connect to database")
	//db, err := driver.InitSQLite(cfg.BotDb)
	//if err != nil {
	//	log.Fatal().Err(err).Msg("error initializing database")
	//	return
	//}
	//_ = storage.NewGormStorage(db)

	log.Info().Msg("connect to Telegram API")
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading config")
		return
	}

	log.Info().Msg("build to mqtt")
	templService, err := templates.New(cfg.TemplateDir)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading templates")
	}
	mqttHandler := NewMqttHandlers(bot, templService, cfg.ChatID)

	opts := mqtt.NewClientOptions().AddBroker(cfg.MqttServer)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	client.Subscribe(cfg.MqttTopic, 1, mqttHandler.statusHandler)
	<-ctx.Done()
	log.Warn().Msg("shutting down")
	time.Sleep(time.Second)
}
