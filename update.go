package telejoon

import (
	"context"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type StateUpdate[User any] struct {
	context    context.Context
	State      string
	User       User
	language   *Language
	Update     tgbotapi.Update
	IsSwitched bool
}

// Set sets a value for the context.
func (s *StateUpdate[User]) Set(key, value interface{}) {
	s.context.Value("storage").(*sync.Map).Store(key, value)
}

// Get gets a value from the context.
func (s *StateUpdate[User]) Get(key interface{}) interface{} {
	val, _ := s.context.Value("storage").(*sync.Map).Load(key)
	return val
}

// SetLanguage sets the language for the user.
func (s *StateUpdate[User]) SetLanguage(language *Language) {
	s.language = language
}

func (s *StateUpdate[User]) Language() *Language {
	return s.language
}
