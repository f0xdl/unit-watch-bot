package tgservice

import (
	"context"
	"github.com/f0xdl/unit-watch-bot/internal/storage"
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	bot       *tgbotapi.BotAPI
	storage   *storage.GormStorage
	templater *templates.TemplateService
	fsm       *FSM
}

func NewHandler(bot *tgbotapi.BotAPI, s *storage.GormStorage, t *templates.TemplateService, fsm *FSM) *Handler {
	return &Handler{bot, s, t, fsm}
}

func (h *Handler) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if update.Message.IsCommand() {
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		switch cmd {
		case "status":
			h.respondWithStatus(chatID, args)
		case "lang":
			h.changeLanguage(chatID, args)
		}
		return
	}

	switch h.fsm.Get(chatID) {
	case StateWaitingUUID:
		h.respondWithStatus(chatID, text)
		h.fsm.Set(chatID, StateIdle)
	}
}

func (h *Handler) respondWithStatus(chatID int64, uuid string) {
	ctx := context.Background()
	device, err := h.storage.Get(ctx, uuid)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, err.Error())) //TODO: fix write error to bot
		return
	}
	if device.UUID == "" {
		h.bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Пристрій не знайдено"))
		return
	}

	args := map[string]interface{}{
		"UUID":     device.UUID,
		"Status":   device.Status.String(),
		"Online":   device.Online,
		"Label":    device.Label,
		"Point":    device.Point,
		"LastSeen": templates.FormatSeenAgo(device.LastSeen, "uk"),
	}

	for k, v := range args {
		switch v.(type) {
		case string:
			args[k] = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, v.(string))
		}
	}
	msg, err := h.templater.Render("device-status.tmpl", args)
	tMsg := tgbotapi.NewMessage(chatID, msg)

	if err != nil {
		tMsg = tgbotapi.NewMessage(chatID, "❌ Помилка при формуванні повідомлення")
		return
	}
	tMsg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err = h.bot.Send(tMsg)
	if err != nil {
		log.Error().Err(err).Msg("error sending message")
	}
}

func (h *Handler) changeLanguage(chatId int64, args string) {
	//ctx := context.Background()
	//TODO add lang to group object in db (chatId+lang)
}
