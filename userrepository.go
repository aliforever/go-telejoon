package telejoon

import (
	"github.com/aliforever/go-telegram-bot-api/structs"
	"sync"
)

type UserRepository interface {
	Upsert(user *structs.User) error
	SetState(id int64, state string) error
	GetState(id int64) (string, error)
}

type UserI[T any] interface {
	FromTgUser(tgUser *structs.User) T
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

func (u *defaultUserRepository) Upsert(user *structs.User) error {
	u.users.Store(user.Id, user)

	return nil
}

func (u *defaultUserRepository) SetState(id int64, state string) error {
	u.states.Store(id, state)
	return nil
}

func (u *defaultUserRepository) GetState(id int64) (string, error) {
	if state, ok := u.states.Load(id); ok {
		return state.(string), nil
	}

	return "", nil
}
