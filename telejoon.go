package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
)

func Start(client *tgbotapi.TelegramBot, processor Processor) {
	for update := range client.Updates() {
		if processor.canProcess(update) {
			go processor.Process(client, update)
		}
	}
}
