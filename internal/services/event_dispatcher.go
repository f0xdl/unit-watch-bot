package services

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type Event struct {
	Type    string
	Update  tgbotapi.Update
	Context context.Context
}

type EventDispatcher struct {
	bot         *tgbotapi.BotAPI
	eventChan   chan Event
	workerCount int
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewEventDispatcher(bot *tgbotapi.BotAPI, workerCount int) *EventDispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &EventDispatcher{
		bot:         bot,
		eventChan:   make(chan Event, 100),
		workerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (ed *EventDispatcher) Start() {
	log.Info().Int("workers", ed.workerCount).Msg("starting event dispatcher")

	log.Warn().Msg("DISPATCHER - IMPLEMENT ME")

	log.Info().Msg("event dispatcher started")
}

func (ed *EventDispatcher) Stop() {
	log.Info().Msg("stopping event dispatcher")
	ed.bot.StopReceivingUpdates()
	ed.cancel()
	close(ed.eventChan)
	log.Info().Msg("event dispatcher stopped")
}
