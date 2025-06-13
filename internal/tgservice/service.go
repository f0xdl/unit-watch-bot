package tgservice

import (
	"github.com/f0xdl/unit-watch-bot/internal/storage"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	handler *Handler
	bot     *tgbotapi.BotAPI
}

func NewService(bot *tgbotapi.BotAPI, store *storage.GormStorage, tmpl *templates.TemplateService) *Service {
	fsm := NewFSM()
	handler := NewHandler(bot, store, tmpl, fsm)
	return &Service{handler, bot}
}

func (s *Service) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := s.bot.GetUpdatesChan(u)
	for update := range updates {
		s.handler.Handle(update)
	}
}
