package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
	"sync"
)

// StaticMenu is a handler that receives predefined handlers and acts accordingly.
type StaticMenu[User any] struct {
	lock sync.Mutex

	replyText            string
	replyWithLanguageKey string

	replyWithFunc func(*tgbotapi.TelegramBot, *StateUpdate[User])

	buttonInlineMenus map[string]string
	buttonTexts       map[string]string
	buttonStates      map[string]string
	buttonFuncs       map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User]) string

	middlewares []func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)

	buttons []string

	actions map[string]bool
}

// NewStaticMenu creates a new raw StaticMenu[User UserI[User]].
func NewStaticMenu[User any]() *StaticMenu[User] {
	return &StaticMenu[User]{
		buttonInlineMenus: make(map[string]string),
		buttonStates:      make(map[string]string),
		buttonFuncs:       make(map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User]) string),
		buttonTexts:       make(map[string]string),
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

// AddButtonInlineMenu adds a new reply button inline menu to the handler.
func (s *StaticMenu[User]) AddButtonInlineMenu(button, menu string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonInlineMenus[button] = menu

	s.buttons = append(s.buttons, button)

	return s
}

// AddCommandInlineMenu adds a new reply button inline menu to the handler.
func (s *StaticMenu[User]) AddCommandInlineMenu(button, menu string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonInlineMenus[button] = menu

	return s
}

// AddButtonText adds a new reply button text to the handler.
func (s *StaticMenu[User]) AddButtonText(button, text string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonTexts[button] = text

	s.buttons = append(s.buttons, button)

	return s
}

// AddButtonState adds a new reply button state to the handler.
func (s *StaticMenu[User]) AddButtonState(button, state string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonStates[button] = state

	s.buttons = append(s.buttons, button)

	return s
}

// AddButtonFunc adds a new reply button func to the handler.
func (s *StaticMenu[User]) AddButtonFunc(
	button string, f func(*tgbotapi.TelegramBot, *StateUpdate[User]) string) *StaticMenu[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonFuncs[button] = f

	s.buttons = append(s.buttons, button)

	return s
}

// AddCommandText adds a new reply button text to the handler.
func (s *StaticMenu[User]) AddCommandText(button, text string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonTexts[button] = text

	return s
}

// AddCommandState adds a new reply button state to the handler.
func (s *StaticMenu[User]) AddCommandState(button, state string) *StaticMenu[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonStates[button] = state

	return s
}

// AddCommandFunc adds a new reply button func to the handler.
func (s *StaticMenu[User]) AddCommandFunc(
	button string, f func(*tgbotapi.TelegramBot, *StateUpdate[User]) string) *StaticMenu[User] {

	s.lock.Lock()
	defer s.lock.Unlock()

	s.buttonFuncs[button] = f

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

func (s *StaticMenu[User]) buildButtonKeyboard() *structs.ReplyKeyboardMarkup {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(s.buttons) == 0 {
		return nil
	}

	return tools.Keyboards{}.NewReplyKeyboardFromSliceOfStrings(s.buttons, 2)
}

func (s *StaticMenu[User]) getReplyTextForButton(button string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonTexts[button]
}

func (s *StaticMenu[User]) getStateForButton(button string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonStates[button]
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

func (s *StaticMenu[User]) getFuncForButton(btn string) func(*tgbotapi.TelegramBot, *StateUpdate[User]) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonFuncs[btn]
}

func (s *StaticMenu[User]) getInlineMenuForButton(btn string) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.buttonInlineMenus[btn]
}
