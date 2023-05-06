package telejoon_test

import (
	"context"
	"fmt"
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

	defaultUserRepo := telejoon.NewDefaultUserRepository[ExampleUser]()

	type args struct {
		client    *tgbotapi.TelegramBot
		processor func() *telejoon.EngineWithPrivateStateHandlers[ExampleUser]
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
				processor: func() *telejoon.EngineWithPrivateStateHandlers[ExampleUser] {
					return telejoon.WithPrivateStateHandlers[ExampleUser](
						defaultUserRepo,
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
								WithDynamicHandlers(telejoon.NewDynamicHandlers[ExampleUser]().
									WithTextHandler(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser]) (string, bool) {
										if update.Update.Message.Text == "Hello Bro" {
											client.Send(client.Message().SetChatId(update.User.Id).
												SetText("Hello Bro!"))
											fmt.Println("update inside dynamic text handler", update)

											fmt.Println("changing name to:", "Ali 2")
											update.Set("name", "Ali 2")

											return "", false
										}

										return "", true
									})).
								ReplyWithLanguageKey("Welcome.Main")).
						AddStaticMenu("Info", telejoon.NewStaticMenu[ExampleUser]().
							WithStaticActionBuilder(telejoon.NewActionBuilder().
								AddStateButtonT("Global.Back", "Welcome")).
							ReplyWithLanguageKey("Info.Hello")).
						AddInlineMenu("Info", telejoon.NewInlineMenu[ExampleUser]().
							WithInlineActionBuilder(telejoon.NewInlineActionBuilder().
								AddUrlButton("Google", "https://google.com").
								AddAlertButtonT("Info.Hello", "say_hello_0", "HI!").
								AddAlertButton("Hello", "say_hello", "Hello Friend").
								AddAlertButton("Hello 2", "say_hello_2", "Hello Friend 2").
								AddAlertButton("Hello 3", "say_hello_3", "Hello Friend 3").
								AddCallbackButton("Callback 1", "callback_1:data").
								AddCallbackButton("Callback 2", "callback_1:data2").
								AddInlineMenuButtonWithEdit("Change Menu to Info 2", "Info2", "Info2").
								SetMaxButtonPerRow(3)).
							WithReplyText("Info Inline Menu")).
						AddInlineMenu("Info2", telejoon.NewInlineMenu[ExampleUser]().
							WithInlineActionBuilder(telejoon.NewInlineActionBuilder().
								AddAlertButtonWithDialog("Hello", "say_hello_4", "Hello Friend").
								AddInlineMenuButtonWithEdit("Back", "Info", "Info")).
							WithReplyText("Info2 Inline Menu")).
						AddCallbackQueryHandler("callback_1", func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate[ExampleUser], args ...string) {
							text := "Callback 1 Clicked"
							if len(args) > 0 {
								text = fmt.Sprintf("Callback 1 Clicked with args: %s", args[0])
							}
							client.Send(client.AnswerCallbackQuery().
								SetCallbackQueryId(update.Update.CallbackQuery.Id).
								SetText(text))
							return
						})
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for update := range tt.args.client.Updates() {
				tt.args.processor().Process(tt.args.client, update)
			}
		})
	}

	<-stop
}
