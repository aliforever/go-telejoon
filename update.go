package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type StateUpdate[User UserI, Lang LanguageI] struct {
	User       User
	Language   Lang
	Update     tgbotapi.Update
	IsSwitched bool
}
