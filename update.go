package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type StateUpdate[User any] struct {
	storage    *sync.Map
	State      string
	User       User
	language   *Language
	Update     tgbotapi.Update
	IsSwitched bool
}

// Set sets a value for the context.
func (s *StateUpdate[User]) Set(key, value interface{}) {
	s.storage.Store(key, value)
}

// Get gets a value from the context.
func (s *StateUpdate[User]) Get(key interface{}) interface{} {
	value, _ := s.storage.Load(key)

	return value
}

// SetLanguage sets the language for the user.
func (s *StateUpdate[User]) SetLanguage(language *Language) {
	s.language = language
}

func (s *StateUpdate[User]) Language() *Language {
	return s.language
}
