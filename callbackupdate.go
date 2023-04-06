package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type CallbackUpdate[User UserI, Language LanguageI] struct {
	User     User
	Language Language
	Update   tgbotapi.Update
}
