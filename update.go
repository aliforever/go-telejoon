package telejoon

import (
	"context"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
)

type StateUpdate[User any] struct {
	context    context.Context
	State      string
	User       User
	Update     tgbotapi.Update
	IsSwitched bool
}

// Set sets a value for the context.
func (s *StateUpdate[User]) Set(key, value interface{}) {
	s.context = context.WithValue(s.context, key, value)
}

// Get gets a value from the context.
func (s *StateUpdate[User]) Get(key interface{}) interface{} {
	return s.context.Value(key)
}

/*type StateUpdateWithLanguage[User UserI[User], Lang LanguageI] struct {
	State      string
	User       User
	Language   Lang
	Update     tgbotapi.Update
	IsSwitched bool
}*/
