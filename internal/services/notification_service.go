package services

import (
	"github.com/f0xdl/unit-watch-bot/internal/templates"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"time"
)

type Notification struct {
	bot       *tgbotapi.BotAPI
	templater *templates.TemplateService
}

func NewNotificationService(bot *tgbotapi.BotAPI, templater *templates.TemplateService) *Notification {
	return &Notification{
		bot:       bot,
		templater: templater,
	}
}

// SendToChats success if 1 or more messages sending
// templateName -- with .tmpl
// chatIds -- map[chatId]LangCode
func (ns *Notification) SendToChats(templateName string, chatIds map[int64]string, args map[string]interface{}) map[int64]error {
	messages := ns.prepareMessages(templateName, chatIds, args)
	errs := map[int64]error{}
	for chatId, msg := range messages {
		msg := tgbotapi.NewMessage(chatId, msg)
		msg.ParseMode = tgbotapi.ModeMarkdownV2

		_, err := ns.bot.Request(msg)
		if err != nil {
			errs[chatId] = err
		}
	}
	log.Debug().
		Int("total", len(messages)).
		Int("errors", len(errs)).
		Msg("status notification sending completed")
	return errs

}

func (ns *Notification) prepareMessages(templateName string, chatIds map[int64]string, args map[string]interface{}) map[int64]string {
	result := map[int64]string{}
	var err error
	for chatId, lang := range chatIds {
		if !templates.LangCodeIsValid(lang) {
			log.Warn().Int64("chat", chatId).Str("lang", lang).Msg("invalid language code")
			lang = string(templates.English)
		}
		prepareArgs(templates.LangCode(lang), args)
		result[chatId], err = ns.templater.Render(templateName, args)
		if err != nil {
			log.Error().Err(err).Int64("chatId", chatId).Msg("render error")
		}
	}
	return result
}

func prepareArgs(lang templates.LangCode, args map[string]interface{}) {
	args["ChangedAt"] = templates.FormatChangedAt(args["ChangedAt"].(time.Time), lang)
	args["LastSeen"] = templates.FormatSeenAgo(args["LastSeen"].(time.Time), lang)

	for k, v := range args {
		if str, ok := v.(string); ok {
			args[k] = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, str)
		}
	}
}
