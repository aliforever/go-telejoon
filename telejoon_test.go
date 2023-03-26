package telejoon_test

import (
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telejoon"
	"os"
	"strings"
	"testing"
)

type User struct {
	Id int64
}

// UserRepository is a dummy implementation of UserRepository interface.
type UserRepository struct{}

func (u *UserRepository) Store(user *structs.User) (*User, error) {
	return &User{Id: user.Id}, nil
}

func (u *UserRepository) Find(id int64) (*User, error) {
	return &User{Id: id}, nil
}

// UserStateRepository is a dummy implementation of UserStateRepository interface.
type UserStateRepository struct{}

func (u *UserStateRepository) Store(id int64, state string) error {
	return nil
}

func (u *UserStateRepository) Find(id int64) (string, error) {
	return "Welcome", nil
}

func TestNew(t *testing.T) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		t.Skip("BOT_TOKEN is not set")
	}

	c, _ := tgbotapi.NewTelegramBot(botToken)

	ch := telejoon.NewCallbackHandlers[User](":").
		AddHandler("info", func(user *User, update tgbotapi.Update, args ...string) {
			c.Send(c.Message().SetChatId(user.Id).SetText("info:" + strings.Join(args, " ")))
		})

	sh := telejoon.NewStateHandlers[User]("Welcome", &UserRepository{}, &UserStateRepository{}).
		AddHandler("Welcome", func(user *User, update tgbotapi.Update, isSwitched bool) string {
			if !isSwitched {
				if update.Message.Text == "/info" {
					return "Info"
				}
				if update.Message.Text == "/callback" {
					c.Send(c.Message().
						SetChatId(user.Id).
						SetText("Choose an option...").
						SetReplyMarkup(c.Tools.Keyboards.NewInlineKeyboardFromSlicesOfMaps([][]map[string]string{
							{{"text": "info:1", "callback_data": "info:1"}},
						})))

					return ""
				}
			}

			c.Send(c.Message().
				SetChatId(user.Id).
				SetText("Choose an option...").
				SetReplyMarkup(c.Tools.Keyboards.NewReplyKeyboardFromSlicesOfStrings([][]string{{"/info"}})))

			return ""
		}).
		AddHandler("Info", func(user *User, update tgbotapi.Update, isSwitched bool) string {
			c.Send(c.Message().SetChatId(user.Id).SetText("Info Menu!"))

			return ""
		})

	d := telejoon.New[User](
		c.GetUpdates().LongPoll(),
		telejoon.NewHandlers[User]().SetStateHandlers(sh).SetCallbackHandlers(ch),
		telejoon.NewOptions().OnErr(onErr),
	)

	d.Start()
}

func onErr(update tgbotapi.Update, err error) {
	fmt.Println(err)
}
