package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

var defaultCommandSeparator = ":"

// CallbackHandlers is a struct that holds all the callback handlers
type CallbackHandlers[T any] struct {
	m sync.Mutex
	// CommandSeparator is the separator between the command and the arguments
	//  for example: add_user:ali:1234 will be separated into:
	//  command: add_user
	//  arguments: ali:1234
	//  if the separator is not set, the default separator will be used
	commandSeparator string

	CallbackHandlers map[string]func(user *T, update tgbotapi.Update, args ...string)
}

// NewCallbackHandlers creates a new instance of CallbackHandlers
func NewCallbackHandlers[T any](separator string) *CallbackHandlers[T] {
	if separator == "" {
		separator = defaultCommandSeparator
	}

	return &CallbackHandlers[T]{
		CallbackHandlers: make(map[string]func(user *T, update tgbotapi.Update, args ...string)),
		commandSeparator: separator,
	}
}

// AddHandler adds a new handler to the CallbackHandlers
func (c *CallbackHandlers[T]) AddHandler(
	command string, handler func(*T, tgbotapi.Update, ...string)) *CallbackHandlers[T] {

	c.m.Lock()
	defer c.m.Unlock()

	c.CallbackHandlers[command] = handler

	return c
}

// GetHandler returns the handler for the given command
func (c *CallbackHandlers[T]) GetHandler(command string) func(*T, tgbotapi.Update, ...string) {
	c.m.Lock()
	defer c.m.Unlock()

	return c.CallbackHandlers[command]
}
