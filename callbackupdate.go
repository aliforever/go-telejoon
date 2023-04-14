package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type CallbackUpdate[User any] struct {
	User   User
	Update tgbotapi.Update
}
