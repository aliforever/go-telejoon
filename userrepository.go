package telejoon

import "github.com/aliforever/go-telegram-bot-api/structs"

type UserRepository[T UserI] interface {
	Store(user *structs.User) (T, error)
	Find(id int64) (T, error)
	SetLanguage(id int64, language string) error
}
