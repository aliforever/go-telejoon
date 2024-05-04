package telejoon

import (
	"github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type StateUpdate struct {
	storage *sync.Map

	State      string
	language   *Language
	Update     tgbotapi.Update
	IsSwitched bool
}

// Set sets a value for the context.
func (s *StateUpdate) Set(key, value interface{}) {
	s.storage.Store(key, value)
}

// Get gets a value from the context.
func (s *StateUpdate) Get(key interface{}) interface{} {
	value, _ := s.storage.Load(key)

	return value
}

// SetLanguage sets the language for the user.
func (s *StateUpdate) SetLanguage(language *Language) {
	s.language = language
}

func (s *StateUpdate) Language() *Language {
	return s.language
}
