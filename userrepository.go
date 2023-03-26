package telejoon

import "github.com/aliforever/go-telegram-bot-api/structs"

type UserRepository[T any] interface {
	Store(user *structs.User) (*T, error)
	Find(id int64) (*T, error)
}
