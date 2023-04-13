package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

var defaultCommandSeparator = ":"

// CallbackHandlers is a struct that holds all the callback handlers
type CallbackHandlers[User UserI[User]] struct {
	m sync.Mutex
	// CommandSeparator is the separator between the command and the arguments
	//  for example: add_user:ali:1234 will be separated into:
	//  command: add_user
	//  arguments: ali:1234
	//  if the separator is not set, the default separator will be used
	commandSeparator string

	CallbackHandlers map[string]func(*tgbotapi.TelegramBot, CallbackUpdate[User], ...string)
}

// NewCallbackHandlers creates a new instance of CallbackHandlers
func NewCallbackHandlers[User UserI[User]](separator string) *CallbackHandlers[User] {
	if separator == "" {
		separator = defaultCommandSeparator
	}

	return &CallbackHandlers[User]{
		CallbackHandlers: make(map[string]func(*tgbotapi.TelegramBot, CallbackUpdate[User], ...string)),
		commandSeparator: separator,
	}
}

// AddHandler adds a new handler to the CallbackHandlers
func (c *CallbackHandlers[User]) AddHandler(
	command string,
	handler func(*tgbotapi.TelegramBot, CallbackUpdate[User], ...string)) *CallbackHandlers[User] {

	c.m.Lock()
	defer c.m.Unlock()

	c.CallbackHandlers[command] = handler

	return c
}

// GetHandler returns the handler for the given command
func (c *CallbackHandlers[User]) GetHandler(command string) func(
	*tgbotapi.TelegramBot, CallbackUpdate[User], ...string) {

	c.m.Lock()
	defer c.m.Unlock()

	return c.CallbackHandlers[command]
}
