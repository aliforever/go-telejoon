package telejoon_test

import (
	"context"
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telejoon"
	"os"
	"testing"
)

type ExampleUser struct {
	Id int64
}

func (e ExampleUser) FromTgUser(tgUser *structs.User) ExampleUser {
	return ExampleUser{tgUser.Id}
}

func (e ExampleUser) LanguageCode() string {
	return "fa"
}

func TestStart(t *testing.T) {
	var stop = make(chan bool)

	client1 := func() *tgbotapi.TelegramBot {
		botToken := os.Getenv("BOT_TOKEN")
		if botToken == "" {
			t.Skip("BOT_TOKEN is not set")
		}

		c, err := tgbotapi.New(botToken)
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			err := c.GetUpdates().LongPoll()
			if err != nil {
				panic(err)
			}
		}()

		return c
	}()

	type args struct {
		client    *tgbotapi.TelegramBot
		processor telejoon.Processor
		context   context.Context
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "TestStart",
			args: args{
				client: client1,

				processor: telejoon.WithPrivateStateHandlers[ExampleUser](
					telejoon.NewDefaultUserRepository[ExampleUser](),
					"Welcome",
					telejoon.NewOptions().SetErrorGroupID(81997375)).
					AddStaticHandler("Welcome",
						telejoon.NewStaticStateHandler[ExampleUser]().
							AddMiddleware(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser]) (string, bool) {
								update.Set("name", "Ali")
								return "", true
							}).
							ReplyWithText("This is Welcome Menu!").
							AddButtonText("Hello", "You said Hello").
							AddButtonText("Bye", "You said Bye").
							AddButtonState("Show Info", "Info").
							AddButtonFunc("Dynamic Info",
								func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser]) string {
									client.Send(client.Message().
										SetChatId(update.User.Id).
										SetText(fmt.Sprintf("Hello %s\nContex Value: %s\nId: %d",
											update.Get("name").(string), update.Get("test").(string),
											update.User.Id)))
									return ""
								}),
					).
					AddStaticHandler("Info",
						telejoon.NewStaticStateHandler[ExampleUser]().
							AddButtonState("Back", "Welcome").
							ReplyWithText("This is Info Menu!").
							ReplyWithFunc(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser]) {
								client.Send(client.Message().
									SetText("replied with func").
									SetChatId(update.User.Id))
							})),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telejoon.Start(tt.args.client, tt.args.processor)
		})
	}

	<-stop
}
