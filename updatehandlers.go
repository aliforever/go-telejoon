package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type Handler interface {
	Handle(client *tgbotapi.TelegramBot, update *StateUpdate) (SwitchAction, bool)
}

type UpdateHandler func(client *tgbotapi.TelegramBot, update *StateUpdate) (SwitchAction, bool)

func (h UpdateHandler) Handle(client *tgbotapi.TelegramBot, update *StateUpdate) (SwitchAction, bool) {
	return h(client, update)
}

type Middleware struct {
	UpdateHandler
}

// NewMiddleware returns a new Middleware that calls Handler.
func NewMiddleware(handler UpdateHandler) Middleware {
	return Middleware{handler}
}

type PanicHandler func(
	client *tgbotapi.TelegramBot,
	update tgbotapi.Update,
	err interface{},
	trace string,
) (SwitchAction, bool)
