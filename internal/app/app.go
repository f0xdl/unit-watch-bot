package app

import (
	"context"
	"github.com/f0xdl/unit-watch-bot/internal/handlers"
	"github.com/f0xdl/unit-watch-bot/internal/services"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	"github.com/f0xdl/unit-watch-lib/configuration"
	"github.com/f0xdl/unit-watch-lib/mqtt_manager"
	"github.com/f0xdl/unit-watch-lib/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"time"
)

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN"`
	BotDb         string `env:"BOT_DB"`
	TemplateDir   string `env:"TEMPLATE_DIR"`
	//MQTT
	MqttServer   string `env:"MQTT_SERVER"`
	MqttClientId string `env:"MQTT_CLIENT_ID"`
	MqttUsername string `env:"MQTT_USERNAME"`
	MqttPassword string `env:"MQTT_PASSWORD"`
}

type App struct {
	cfg                Config
	mqtt               *mqtt_manager.MqttManager
	telegramDispatcher *services.EventDispatcher
}

func SetupApp() (a *App, err error) {
	//TODO: recovery SetupApp
	a = &App{}

	log.Info().Msg("read configuration")
	a.cfg = Config{}
	err = configuration.LoadEnvConfig(&a.cfg)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("connect to database")
	store, err := repository.NewDeviceRepository(sqlite.Open(a.cfg.BotDb), true)
	if err != nil {
		return nil, err
	}
	deviceService := services.NewDeviceService(store)

	log.Info().Msg("read templates")
	templater, err := templates.New(a.cfg.TemplateDir)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("build telegram bot")
	tbot, err := tgbotapi.NewBotAPI(a.cfg.TelegramToken)
	if err != nil {
		return nil, err
	}
	notification := services.NewNotificationService(tbot, templater)
	a.telegramDispatcher = services.NewEventDispatcher(tbot, 3)

	log.Info().Msg("build mqtt client")
	adapter := mqtt_manager.NewMessageHandlerAdapter
	parser := mqtt_manager.NewDefaultTopicParser()
	topics := map[string]mqtt_manager.MessageHandler{
		"device/+/status": adapter(handlers.StatusChanged(deviceService, notification).Handle, parser),
		"device/+/online": adapter(handlers.DeviceOnline(deviceService, notification).Handle, parser),
	}
	a.mqtt = mqtt_manager.NewMqttManager(a.cfg.MqttServer, a.cfg.MqttClientId, a.cfg.MqttUsername, a.cfg.MqttPassword)
	for topic, handler := range topics {
		a.mqtt.SetTopicHandle(topic, handler)
	}

	return
}

func (app *App) Run(ctx context.Context) (err error) {
	//TODO: recovery Run
	err = app.mqtt.Connect()
	if err != nil {
		return err
	}
	app.telegramDispatcher.Start()

	<-ctx.Done()
	log.Info().Msg("graceful shutdown")
	canceledCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	select {
	case <-app.Shutdown():
		log.Info().Msg("graceful shutdown finished")
	case <-canceledCtx.Done():
		log.Error().Msg("graceful shutdown canceled")
	}
	return
}

func (app *App) Shutdown() chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		app.mqtt.Disconnect(1000)
		app.telegramDispatcher.Stop()
	}()
	return ch
}
