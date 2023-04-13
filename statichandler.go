package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
	"sync"
)

// StaticStateHandler is a handler that receives predefined handlers and acts accordingly.
type StaticStateHandler[User any] struct {
	lock sync.Mutex

	replyText string

	replyWithFunc func(*tgbotapi.TelegramBot, *StateUpdate[User])

	buttonTexts  map[string]string
	buttonStates map[string]string
	buttonFuncs  map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User]) string

	middlewares []func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)

	buttons []string
}

// NewStaticStateHandler creates a new raw StaticStateHandler[User UserI[User]].
func NewStaticStateHandler[User UserI[User]]() *StaticStateHandler[User] {
	return &StaticStateHandler[User]{
		buttonStates: make(map[string]string),
		buttonFuncs:  make(map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User]) string),
		buttonTexts:  make(map[string]string),
	}
}

// AddMiddleware adds a new middleware to the handler.
func (s *StaticStateHandler[User]) AddMiddleware(
	m func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)) *StaticStateHandler[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.middlewares = append(s.middlewares, m)

	return s
}

// AddButtonText adds a new reply button text to the handler.
func (s *StaticStateHandler[User]) AddButtonText(button, text string) *StaticStateHandler[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonTexts[button] = text

	s.buttons = append(s.buttons, button)

	return s
}

// AddButtonState adds a new reply button state to the handler.
func (s *StaticStateHandler[User]) AddButtonState(button, state string) *StaticStateHandler[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonStates[button] = state

	s.buttons = append(s.buttons, button)

	return s
}

// AddButtonFunc adds a new reply button func to the handler.
func (s *StaticStateHandler[User]) AddButtonFunc(
	button string, f func(*tgbotapi.TelegramBot, *StateUpdate[User]) string) *StaticStateHandler[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonFuncs[button] = f

	s.buttons = append(s.buttons, button)

	return s
}

// AddCommandText adds a new reply button text to the handler.
func (s *StaticStateHandler[User]) AddCommandText(button, text string) *StaticStateHandler[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonTexts[button] = text

	return s
}

// AddCommandState adds a new reply button state to the handler.
func (s *StaticStateHandler[User]) AddCommandState(button, state string) *StaticStateHandler[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonStates[button] = state

	return s
}

// AddCommandFunc adds a new reply button func to the handler.
func (s *StaticStateHandler[User]) AddCommandFunc(
	button string, f func(*tgbotapi.TelegramBot, *StateUpdate[User]) string) *StaticStateHandler[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonFuncs[button] = f

	return s
}

func (s *StaticStateHandler[User]) ReplyWithText(text string) *StaticStateHandler[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.replyText = text
	return s
}

func (s *StaticStateHandler[User]) ReplyWithFunc(
	f func(*tgbotapi.TelegramBot, *StateUpdate[User])) *StaticStateHandler[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.replyWithFunc = f
	return s
}

func (s *StaticStateHandler[User]) buildButtonKeyboard() *structs.ReplyKeyboardMarkup {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.buttons) == 0 {
		return nil
	}

	return tools.Keyboards{}.NewReplyKeyboardFromSliceOfStrings(s.buttons, 2)
}

func (s *StaticStateHandler[User]) getReplyTextForButton(button string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonTexts[button]
}

func (s *StaticStateHandler[User]) getStateForButton(button string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonStates[button]
}

func (s *StaticStateHandler[User]) getReplyText() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyText
}

func (s *StaticStateHandler[User]) getReplyWithFunc() func(*tgbotapi.TelegramBot, *StateUpdate[User]) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyWithFunc
}

func (s *StaticStateHandler[User]) getFuncForButton(btn string) func(*tgbotapi.TelegramBot, *StateUpdate[User]) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonFuncs[btn]
}
