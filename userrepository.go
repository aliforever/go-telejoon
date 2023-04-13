package telejoon

import (
	"fmt"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"sync"
)

type UserRepository[T any] interface {
	Store(user *structs.User) (T, error)
	Find(id int64) (T, error)
	SetState(id int64, state string) error
	GetState(id int64) (string, error)
}

type UserI[T any] interface {
	FromTgUser(tgUser *structs.User) T
}

type defaultUserRepository[User UserI[User]] struct {
	users  sync.Map
	states sync.Map
}

// NewDefaultUserRepository Factory function for defaultUserRepository.
func NewDefaultUserRepository[User UserI[User]]() UserRepository[User] {
	return &defaultUserRepository[User]{
		users:  sync.Map{},
		states: sync.Map{},
	}
}

func (u *defaultUserRepository[User]) Store(user *structs.User) (User, error) {
	var us User

	modelUser := us.FromTgUser(user)

	u.users.Store(user.Id, modelUser)

	return modelUser, nil
}

func (u *defaultUserRepository[User]) Find(id int64) (User, error) {
	if user, ok := u.users.Load(id); ok {
		return user.(User), nil
	}

	return *new(User), fmt.Errorf("user not found")
}

func (u *defaultUserRepository[User]) SetState(id int64, state string) error {
	u.states.Store(id, state)
	return nil
}

func (u *defaultUserRepository[User]) GetState(id int64) (string, error) {
	if state, ok := u.states.Load(id); ok {
		return state.(string), nil
	}

	return "", nil
}
