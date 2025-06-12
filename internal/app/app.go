package app

import (
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/f0xdl/unit-watch-bot/internal/configs"
	"github.com/f0xdl/unit-watch-bot/internal/handlers"
	"github.com/f0xdl/unit-watch-bot/internal/storage"
	"github.com/f0xdl/unit-watch-bot/internal/storage/driver"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"time"
)

type App struct {
	cfg   configs.Config
	store *storage.GormStorage
	//templater   *templates.TemplateService
	mqttClient  mqtt.Client
	tbot        *tgbotapi.BotAPI
	mqttHandler *handlers.MqttHandler
}

func SetupApp() (a *App, err error) {
	//TODO: recovery
	log.Info().Msg("read configuration")
	cfg, err := configs.LoadConfig()
	if err != nil {
		return
	}

	log.Info().Msg("connect to database")
	db, err := driver.InitSQLite(cfg.BotDb)
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing database")
		return
	}
	store := storage.NewGormStorage(db)

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

	mqttHandler := handlers.NewMqttHandler(tbot, templater, cfg.ChatID)

	/* TODO: FSM for tbot
	// FSM для управления состояниями
	var botFSM = fsm.NewFSM(
		"initial",
		fsm.Events{
			{Name: "start", Src: []string{"initial"}, Dst: "waiting"},
			{Name: "message_received", Src: []string{"waiting"}, Dst: "processing"},
			{Name: "processed", Src: []string{"processing"}, Dst: "waiting"},
		},
		fsm.Callbacks{},
	)
	*/
	return &App{
		cfg:   cfg,
		store: store,
		//templater:   templater,
		mqttClient:  mqttClient,
		tbot:        tbot,
		mqttHandler: mqttHandler,
	}, nil
}

func (app *App) Run(ctx context.Context) (err error) {
	token := app.mqttClient.Connect()
	token.Wait()
	app.mqttHandler.SubscribeTopics(app.mqttClient)

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
		time.Sleep(time.Second) //TODO: do  graceful shutdown
	}()
	return ch
}
