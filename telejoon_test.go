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

	defaultUserRepo := telejoon.NewDefaultUserRepository()

	type args struct {
		client    *tgbotapi.TelegramBot
		processor func() *telejoon.EngineWithPrivateStateHandlers
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
				processor: func() *telejoon.EngineWithPrivateStateHandlers {
					return telejoon.WithPrivateStateHandlers(
						defaultUserRepo,
						"Welcome",
						telejoon.NewOptions().SetErrorGroupID(81997375)).
						WithLanguageConfig(languageConfig).
						AddMiddleware(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate) (telejoon.SwitchAction, bool) {
							return nil, true
						}).
						AddStaticMenu("Welcome",
							telejoon.NewStaticMenu(
								telejoon.NewLanguageKeyText("Welcome.Main"),
								telejoon.NewStaticActionBuilder().
									AddStateButtonT("Welcome.ChangeLanguageBtn", "ChangeLanguage").
									AddTextButton("Hello", "You said Hello").
									AddStateButton("Info State", "Info").
									AddInlineMenuButton("Info", "Info"),
								telejoon.NewDynamicHandlerText(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate) (telejoon.SwitchAction, bool) {
									if update.Update.Message.Text == "Hello Bro" {
										client.Send(client.Message().SetChatId(update.Update.From().Id).
											SetText("Hello Bro!"))
										fmt.Println("update inside dynamic text Handler", update)

										fmt.Println("changing name to:", "Ali 2")
										update.Set("name", "Ali 2")

										return nil, false
									}

									return nil, true
								}),
								telejoon.NewMiddleware(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate) (telejoon.SwitchAction, bool) {
									update.Set("name", "Ali")

									return nil, true
								},
								))).
						AddStaticMenu("Info", telejoon.NewStaticMenu(
							telejoon.NewLanguageKeyText("Info.Hello"),
							telejoon.NewStaticActionBuilder().
								AddStateButtonT("Global.Back", "Welcome"))).
						AddInlineMenu("Info", telejoon.
							NewInlineMenuWithTextAndActionBuilder("Info Inline Menu",
								telejoon.NewInlineActionBuilder().
									AddUrlButton("Google", "https://google.com").
									AddAlertButtonT("Info.Hello", "say_hello_0", "HI!").
									AddAlertButton("Hello", "say_hello", "Hello Friend").
									AddAlertButton("Hello 2", "say_hello_2", "Hello Friend 2").
									AddAlertButton("Hello 3", "say_hello_3", "Hello Friend 3").
									AddCallbackButton("Callback 1", "callback_1:data").
									AddCallbackButton("Callback 2", "callback_1:data2").
									AddInlineMenuButtonWithEdit("Change Menu to Info 2", "Info2", "Info2").
									SetMaxButtonPerRow(3))).
						AddInlineMenu("Info2", telejoon.
							NewInlineMenuWithTextAndActionBuilder(
								"Info2 Inline Menu", telejoon.NewInlineActionBuilder().
									AddAlertButtonWithDialog("Hello", "say_hello_4", "Hello Friend").
									AddInlineMenuButtonWithEdit("CustomInline", "CustomInline", "CustomInline").
									AddInlineMenuButtonWithEdit("Back", "Info", "Info"))).
						AddCallbackQueryHandler("callback_1", callbackHandler).
						AddInlineMenu("CustomInline", CustomInlineMenu())
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

func callbackHandler(
	client *tgbotapi.TelegramBot,
	update *telejoon.StateUpdate,
	args ...string,
) (telejoon.SwitchAction, error) {

	text := "Callback 1 Clicked"
	if len(args) > 0 {
		text = fmt.Sprintf("Callback 1 Clicked with args: %s", args[0])
	}
	client.Send(client.AnswerCallbackQuery().
		SetCallbackQueryId(update.Update.CallbackQuery.Id).
		SetText(text))
	return nil, nil
}

func CustomInlineMenu() *telejoon.InlineMenu {
	deferredBuilder := func(update *telejoon.StateUpdate) *telejoon.InlineActionBuilder {
		return telejoon.NewInlineActionBuilder().
			AddAlertButtonWithDialog("Hello", "say_hello_4", "Hello Friend")
	}

	return telejoon.
		NewInlineMenuWithTextAndDeferredActionBuilder("Custom Inline Menu", deferredBuilder)
}
