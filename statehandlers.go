package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type StateHandlers[T any] struct {
	m sync.Mutex

	defaultState string

	userRepository      UserRepository[T]
	userStateRepository UserStateRepository

	handlers map[string]func(user *T, update tgbotapi.Update, isSwitched bool) string
}

func NewStateHandlers[T any](
	defaultState string, userRepo UserRepository[T], userStateRepo UserStateRepository) *StateHandlers[T] {

	return &StateHandlers[T]{
		defaultState:        defaultState,
		userRepository:      userRepo,
		userStateRepository: userStateRepo,
		handlers:            make(map[string]func(user *T, update tgbotapi.Update, isSwitched bool) string),
	}
}

func (s *StateHandlers[T]) AddHandler(
	state string, handler func(user *T, update tgbotapi.Update, isSwitched bool) string) *StateHandlers[T] {

	s.m.Lock()
	defer s.m.Unlock()

	s.handlers[state] = handler

	return s
}

func (s *StateHandlers[T]) GetHandler(state string) func(*T, tgbotapi.Update, bool) string {
	s.m.Lock()
	defer s.m.Unlock()

	return s.handlers[state]
}
