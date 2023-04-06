package telejoon

import (
	"sync"
)

type StateHandlers[User UserI, Lang LanguageI] struct {
	m sync.Mutex

	defaultState string

	userRepository      UserRepository[User]
	userStateRepository UserStateRepository

	handlers map[string]func(update StateUpdate[User, Lang]) string
}

func NewStateHandlers[User UserI, Lang LanguageI](
	defaultState string, userRepo UserRepository[User], userStateRepo UserStateRepository) *StateHandlers[User, Lang] {

	return &StateHandlers[User, Lang]{
		defaultState:        defaultState,
		userRepository:      userRepo,
		userStateRepository: userStateRepo,
		handlers:            make(map[string]func(update StateUpdate[User, Lang]) string),
	}
}

func (s *StateHandlers[User, Lang]) AddHandler(
	state string, handler func(update StateUpdate[User, Lang]) string) *StateHandlers[User, Lang] {

	s.m.Lock()
	defer s.m.Unlock()

	s.handlers[state] = handler

	return s
}

func (s *StateHandlers[User, Lang]) GetHandler(state string) func(update StateUpdate[User, Lang]) string {
	s.m.Lock()
	defer s.m.Unlock()

	return s.handlers[state]
}
