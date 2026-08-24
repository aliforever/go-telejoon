package telejoon

import (
	"context"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

type Processor interface {
	canProcess(update tgbotapi.Update) bool
	Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update)
}
