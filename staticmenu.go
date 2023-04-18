package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

// StaticMenu is a handler that receives predefined handlers and acts accordingly.
type StaticMenu[User any] struct {
	lock sync.Mutex

	replyText            string
	replyWithLanguageKey string

	replyWithFunc func(*tgbotapi.TelegramBot, *StateUpdate[User])

	buttonFuncs        map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User]) string
	languageKeyButtons map[string]bool

	staticActionBuilder *staticActionBuilder

	middlewares []func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)

	buttons []string

	actions map[string]bool
}

// NewStaticMenu creates a new raw StaticMenu[User UserI[User]].
func NewStaticMenu[User any]() *StaticMenu[User] {
	return &StaticMenu[User]{
		buttonFuncs:        make(map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User]) string),
		languageKeyButtons: make(map[string]bool),
	}
}

// AddMiddleware adds a new middleware to the handler.
func (s *StaticMenu[User]) AddMiddleware(
	m func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)) *StaticMenu[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.middlewares = append(s.middlewares, m)

	return s
}

// WithStaticActionBuilder sets the static action builder for the handler.
func (s *StaticMenu[User]) WithStaticActionBuilder(
	builder *staticActionBuilder) *StaticMenu[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.staticActionBuilder = builder

	return s
}

func (s *StaticMenu[User]) ReplyWithText(text string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.replyText = text
	return s
}

func (s *StaticMenu[User]) ReplyWithLanguageKey(key string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.replyWithLanguageKey = key
	return s
}

func (s *StaticMenu[User]) ReplyWithFunc(
	f func(*tgbotapi.TelegramBot, *StateUpdate[User])) *StaticMenu[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.replyWithFunc = f
	return s
}

func (s *StaticMenu[User]) getReplyText() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyText
}

func (s *StaticMenu[User]) getReplyTextLanguageKey() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyWithLanguageKey
}

func (s *StaticMenu[User]) getReplyWithFunc() func(*tgbotapi.TelegramBot, *StateUpdate[User]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyWithFunc
}
