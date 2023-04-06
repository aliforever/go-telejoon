package telejoon

import (
	"sync"
)

var defaultCommandSeparator = ":"

// CallbackHandlers is a struct that holds all the callback handlers
type CallbackHandlers[User UserI, Lang LanguageI] struct {
	m sync.Mutex
	// CommandSeparator is the separator between the command and the arguments
	//  for example: add_user:ali:1234 will be separated into:
	//  command: add_user
	//  arguments: ali:1234
	//  if the separator is not set, the default separator will be used
	commandSeparator string

	CallbackHandlers map[string]func(update CallbackUpdate[User, Lang], args ...string)
}

// NewCallbackHandlers creates a new instance of CallbackHandlers
func NewCallbackHandlers[User UserI, Lang LanguageI](separator string) *CallbackHandlers[User, Lang] {
	if separator == "" {
		separator = defaultCommandSeparator
	}

	return &CallbackHandlers[User, Lang]{
		CallbackHandlers: make(map[string]func(update CallbackUpdate[User, Lang], args ...string)),
		commandSeparator: separator,
	}
}

// AddHandler adds a new handler to the CallbackHandlers
func (c *CallbackHandlers[User, Lang]) AddHandler(
	command string, handler func(CallbackUpdate[User, Lang], ...string)) *CallbackHandlers[User, Lang] {

	c.m.Lock()
	defer c.m.Unlock()

	c.CallbackHandlers[command] = handler

	return c
}

// GetHandler returns the handler for the given command
func (c *CallbackHandlers[User, Lang]) GetHandler(command string) func(CallbackUpdate[User, Lang], ...string) {
	c.m.Lock()
	defer c.m.Unlock()

	return c.CallbackHandlers[command]
}
