package telejoon_test

import (
	"context"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telejoon"
	"golang.org/x/text/language"
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

	languages, err := telejoon.NewLanguageBuilder(language.English).
		RegisterTomlFormat([]string{
			`C:\golang\src\github.com\aliforever\go-telejoon\locale.en.toml`,
			`C:\golang\src\github.com\aliforever\go-telejoon\locale.fa.toml`,
		}).Build()
	if err != nil {
		t.Fatal(err)
	}

	languageConfig := telejoon.NewLanguageConfig(languages, telejoon.NewDefaultUserLanguageRepository()).
		WithChangeLanguageMenu("ChangeLanguage", true)

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
					WithLanguageConfig(languageConfig).
					AddStaticMenu("Welcome",
						telejoon.NewStaticMenu[ExampleUser]().
							AddMiddleware(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser]) (string, bool) {
								update.Set("name", "Ali")
								return "", true
							}).
							WithStaticActionBuilder(telejoon.NewActionBuilder().
								AddStateButtonT("Welcome.ChangeLanguageBtn", "ChangeLanguage").
								AddTextButton("Hello", "You said Hello").
								AddStateButton("Info State", "Info").
								AddInlineMenuButton("Info", "Info")).
							ReplyWithLanguageKey("Welcome.Main")).
					// AddStaticMenu("Info",
					// 	telejoon.NewStaticMenu[ExampleUser]().
					// 		AddButtonState("Back", "Welcome").
					// 		ReplyWithText("This is Info Menu!").
					// 		ReplyWithFunc(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser]) {
					// 			client.Send(client.Message().
					// 				SetText("replied with func").
					// 				SetChatId(update.User.Id))
					// 		})).
					AddStaticMenu("Info", telejoon.NewStaticMenu[ExampleUser]().
						WithStaticActionBuilder(telejoon.NewActionBuilder().
							AddStateButtonT("Global.Back", "Welcome")).
						ReplyWithLanguageKey("Info.Hello")).
					AddInlineMenu("Info", telejoon.NewInlineMenu[ExampleUser]().
						AddButtonUrl("Google", "https://google.com").
						AddLanguageKeyDataButtonAlert("Info.Hello", "say_hello_0", "Hello Friend", false).
						AddDataButtonAlert("Hello", "say_hello", "Hello Friend", false).
						AddDataButtonAlert("Hello 2", "say_hello_2", "Hello Friend 2", false).
						AddDataButtonAlert("Hello 3", "say_hello_3", "Hello Friend 3", true).
						AddButtonInlineMenu("Change Menu to Info 2", "Info2", true).
						SetMaxButtonPerRow(3).
						// SetButtonFormation(1, 3).
						AddReplyText("Info Inline Menu")).
					AddInlineMenu("Info2", telejoon.NewInlineMenu[ExampleUser]().
						AddDataButtonAlert("Hello", "say_hello", "Hello Friend", false).
						AddButtonInlineMenu("Back", "Info", true).
						AddReplyText("Info2 Inline Menu")),
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
