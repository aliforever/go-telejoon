package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type Processor interface {
	canProcess(update tgbotapi.Update) bool
	process(client *tgbotapi.TelegramBot, update tgbotapi.Update)
}
