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

func (u User) LanguageCode() string {
	return "fa"
}

// UserRepository is a dummy implementation of UserRepository interface.
type UserRepository struct{}

func (u *UserRepository) Store(user *structs.User) (User, error) {
	return User{Id: user.Id}, nil
}

func (u *UserRepository) Find(id int64) (User, error) {
	return User{Id: id}, nil
}

func (u *UserRepository) SetLanguage(id int64, language string) error {
	return nil
}

// UserStateRepository is a dummy implementation of UserStateRepository interface.
type UserStateRepository struct{}

func (u *UserStateRepository) Update(userID int64, state string) error {
	return nil
}

func (u *UserStateRepository) Store(id int64, state string) error {
	return nil
}

func (u *UserStateRepository) Find(id int64) (string, error) {
	return "Welcome", nil
}

type language interface {
	telejoon.LanguageI
	Welcome() string
}

type Farsi struct {
}

func (f Farsi) Flag() string {
	return "🇮🇷"
}

func (f Farsi) Code() string {
	return "fa"
}

func (f Farsi) Name() string {
	return "Farsi"
}

func (f Farsi) SelectLanguage() string {
	return "انتخاب زبان"
}

func (f Farsi) Welcome() string {
	return "خوش آمدید"
}

func TestNew(t *testing.T) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		t.Skip("BOT_TOKEN is not set")
	}

	c, _ := tgbotapi.New(botToken)

	go func() {
		err := c.GetUpdates().LongPoll()
		if err != nil {
			panic(err)
		}
	}()

	ch := telejoon.NewCallbackHandlers[User, language](":").
		AddHandler("info", func(update telejoon.CallbackUpdate[User, language], args ...string) {
			c.Send(c.Message().SetChatId(update.User.Id).SetText("info:" + strings.Join(args, " ")))
		})

	sh := telejoon.NewStateHandlers[User, language]("Welcome", &UserRepository{}, &UserStateRepository{}).
		AddHandler("Welcome", func(update telejoon.StateUpdate[User, language]) string {
			if !update.IsSwitched {
				if update.Update.Message.Text == "/info" {
					return "Info"
				}
				if update.Update.Message.Text == "/callback" {
					c.Send(c.Message().
						SetChatId(update.User.Id).
						SetText("Choose an option...").
						SetReplyMarkup(c.Tools.Keyboards.NewInlineKeyboardFromSlicesOfMaps([][]map[string]string{
							{{"text": "info:1", "callback_data": "info:1"}},
						})))

					return ""
				}
			}

			c.Send(c.Message().
				SetChatId(update.User.Id).
				SetText("Choose an option...").
				SetReplyMarkup(c.Tools.Keyboards.NewReplyKeyboardFromSlicesOfStrings([][]string{{"/info"}})))

			return ""
		}).
		AddHandler("Info", func(update telejoon.StateUpdate[User, language]) string {
			c.Send(c.Message().SetChatId(update.User.Id).SetText("Info Menu!"))

			return ""
		})

	d := telejoon.New[User, language](
		c,
		telejoon.NewHandlers[User, language]().
			SetStateHandlers(sh).
			SetCallbackHandlers(ch),
		[]language{Farsi{}},
		telejoon.NewOptions().OnErr(onErr),
	)

	d.Start()
}

func onErr(update tgbotapi.Update, err error) {
	fmt.Println(err)
}
