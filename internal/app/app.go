package app

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-bot/internal/handlers"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	"github.com/f0xdl/unit-watch-bot/internal/tgservice"
	"github.com/f0xdl/unit-watch-lib/configuration"
	"github.com/f0xdl/unit-watch-lib/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"time"
)

type Config struct {
	TelegramToken string `env:"TELEGRAM_TOKEN"`
	MqttServer    string `env:"MQTT_SERVER"`
	MqttTopic     string `env:"MQTT_TOPIC"`
	BotDb         string `env:"BOT_DB"`
	TemplateDir   string `env:"TEMPLATE_DIR"`
}

type App struct {
	cfg         Config
	mqttClient  mqtt.Client
	tbot        *tgbotapi.BotAPI
	mqttHandler *handlers.MqttHandler
	tgService   *tgservice.Service
}

func SetupApp() (a *App, err error) {
	//TODO: recovery
	log.Info().Msg("read configuration")
	cfg := &Config{}
	err = configuration.LoadEnvConfig(cfg)
	if err != nil {
		return
	}

	log.Info().Msg("connect to database")
	store, err := storage.NewGormStorage(sqlite.Open(cfg.BotDb), true)
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing database")
		return
	}

	log.Info().Msg("read templates")
	templater, err := templates.New(cfg.TemplateDir)
	if err != nil {
		log.Fatal().Err(err).Msg("error loading templates")
	}

	log.Info().Msg("build mqtt client")
	opts := mqtt.NewClientOptions().AddBroker(cfg.MqttServer)
	mqttClient := mqtt.NewClient(opts)

	log.Info().Msg("build telegram bot")
	tbot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, err
	}
	tgService := tgservice.NewService(tbot, store, templater)

	mqttHandler := handlers.NewMqttHandler(tbot, templater, store)

	return &App{
		mqttClient:  mqttClient,
		tbot:        tbot,
		mqttHandler: mqttHandler,
		tgService:   tgService,
	}, nil
}

func (app *App) Run(ctx context.Context) (err error) {
	token := app.mqttClient.Connect()
	token.Wait()
	app.mqttHandler.SubscribeTopics(app.mqttClient)
	go app.tgService.Run()

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
		app.mqttClient.Disconnect(250)
		app.tbot.StopReceivingUpdates()
		time.Sleep(time.Second) //TODO: do  graceful shutdown
	}()
	return ch
}
