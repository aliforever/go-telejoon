package telejoon

import (
	"errors"
	"sync"

	"github.com/aliforever/go-telegram-bot-api/structs"
)

// UserStateNotFoundErr should be returned by UserRepository.GetUserState
// implementations when no state is stored for the user yet. The engine treats
// it as "new user" and initializes the default state; any other error is
// treated as a real repository failure and surfaced.
var UserStateNotFoundErr = errors.New("user_state_not_found")

type UserRepository interface {
	UpsertUser(user *structs.User) error
	SetUserState(id int64, state string) error
	GetUserState(id int64) (string, error)
}

type defaultUserRepository struct {
	users  sync.Map
	states sync.Map
}

// NewDefaultUserRepository Factory function for defaultUserRepository.
func NewDefaultUserRepository() UserRepository {
	return &defaultUserRepository{
		users:  sync.Map{},
		states: sync.Map{},
	}
}

func (u *defaultUserRepository) UpsertUser(user *structs.User) error {
	u.users.Store(user.Id, user)

	return nil
}

func (u *defaultUserRepository) SetUserState(id int64, state string) error {
	u.states.Store(id, state)
	return nil
}

func (u *defaultUserRepository) GetUserState(id int64) (string, error) {
	if state, ok := u.states.Load(id); ok {
		return state.(string), nil
	}

	return "", UserStateNotFoundErr
}
