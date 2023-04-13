package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type CallbackUpdate[User UserI[User]] struct {
	User   User
	Update tgbotapi.Update
}
