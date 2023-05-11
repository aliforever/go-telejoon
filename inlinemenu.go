package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type InlineMenu[User any] struct {
	lock sync.Mutex

	replyText string

	callbackPrefix string

	inlineActionBuilder *inlineActionBuilder

	middlewares []func(*tgbotapi.TelegramBot, *StateUpdate[User]) bool
}

func NewInlineMenu[User any]() *InlineMenu[User] {
	return &InlineMenu[User]{}
}

// AddMiddleware adds a middleware to the inline menu
func (i *InlineMenu[User]) AddMiddleware(middleware func(*tgbotapi.TelegramBot, *StateUpdate[User]) bool) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.middlewares = append(i.middlewares, middleware)

	return i
}

func (i *InlineMenu[User]) WithReplyText(text string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.replyText = text

	return i
}

func (i *InlineMenu[User]) WithInlineActionBuilder(
	builder *inlineActionBuilder) *InlineMenu[User] {

	i.lock.Lock()
	defer i.lock.Unlock()

	i.inlineActionBuilder = builder

	return i
}

// getReplyText returns the reply text.
func (i *InlineMenu[User]) getReplyText() string {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.replyText
}

// getMiddlewares returns the middlewares.
func (i *InlineMenu[User]) getMiddlewares() []func(bot *tgbotapi.TelegramBot, update *StateUpdate[User]) bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.middlewares
}
