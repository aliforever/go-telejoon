package telejoon_test

import (
	"context"
	"encoding/json"
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
						WithPanicHandler(func(client *tgbotapi.TelegramBot, update tgbotapi.Update, err any, stack string) {
							fmt.Println("Panic Handler", update, "\n", stack)
						}).
						AddMiddleware(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate) (telejoon.SwitchAction, bool) {
							if update.Update.Message.Text == "panic" {
								panic("Panic Test")
							}
							return nil, true
						}).
						AddStaticMenu("Welcome",
							telejoon.NewStaticMenu(
								telejoon.NewLanguageKeyText("Welcome.Main"),
								telejoon.NewStaticActionBuilder().
									AddStateButton(telejoon.NewLanguageKeyText("Welcome.ChangeLanguageBtn"), "ChangeLanguage").
									AddTextButton(telejoon.NewStaticText("Hello"), telejoon.NewStaticText("You said Hello")).
									AddStateButton(telejoon.NewStaticText("Info State"), "Info").
									AddInlineMenuButton(telejoon.NewStaticText("Info"), "Info"),
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
								AddStateButton(telejoon.NewLanguageKeyText("Global.Back"), "Welcome"))).
						AddInlineMenu("Info", telejoon.
							NewInlineMenu(telejoon.NewStaticText("Info Inline Menu"),
								telejoon.NewInlineActionBuilder().
									AddUrlButton(telejoon.NewStaticText("Google"), telejoon.NewStaticText("https://google.com")).
									AddAlertButton(telejoon.NewLanguageKeyText("Info.Hello"), telejoon.NewStaticText("say_hello_0"), "HI!").
									AddAlertButton(telejoon.NewStaticText("Hello"), telejoon.NewStaticText("say_hello"), "Hello Friend").
									AddAlertButton(telejoon.NewStaticText("Hello 2"), telejoon.NewStaticText("say_hello_2"), "Hello Friend 2").
									AddAlertButton(telejoon.NewStaticText("Hello 3"), telejoon.NewStaticText("say_hello_3"), "Hello Friend 3").
									AddCallbackButton(telejoon.NewStaticText("Callback 1"), telejoon.NewStaticText("callback_1:data"), callbackHandler).
									AddCallbackButton(telejoon.NewStaticText("Callback 2"), telejoon.NewStaticText("callback_1:data2"), callbackHandler).
									AddInlineMenuButtonWithEdit(telejoon.NewStaticText("Change Menu to Info 2"), telejoon.NewStaticText("Info2"), "Info2").
									SetMaxButtonPerRow(3))).
						AddInlineMenu("Info2", telejoon.
							NewInlineMenu(
								telejoon.NewStaticText("Info2 Inline Menu"), telejoon.NewInlineActionBuilder().
									AddAlertButtonWithDialog(telejoon.NewStaticText("Hello"), telejoon.NewStaticText("say_hello_4"), "Hello Friend").
									AddInlineMenuButtonWithEdit(telejoon.NewStaticText("CustomInline"), telejoon.NewStaticText("CustomInline"), "CustomInline").
									AddInlineMenuButtonWithEdit(telejoon.NewStaticText("Back"), telejoon.NewStaticText("Info"), "Info"))).
						AddInlineMenu("CustomInline", CustomInlineMenu()).
						AddMiddleware(func(client *tgbotapi.TelegramBot, update *telejoon.StateUpdate) (telejoon.SwitchAction, bool) {
							fmt.Println("update inside middleware", update)

							if update.Update.Message != nil {
								msg, err := client.Send(client.ForwardMessage().
									SetMessageId(update.Update.Message.MessageId).
									SetChatId(-881430497).
									SetFromChatId(update.Update.Message.Chat.Id))
								if err != nil {
									fmt.Println("Error in sending message", err)
								} else if msg != nil {
									j, _ := json.Marshal(update)
									_, err = client.Send(client.Message().
										SetChatId(-881430497).
										SetReplyToMessageId(msg.Message.MessageId).
										SetText(string(j)))
								}
							}

							return nil, true
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
			AddAlertButtonWithDialog(telejoon.NewStaticText("Hello"), telejoon.NewStaticText("say_hello_4"), "Hello Friend")
	}

	return telejoon.
		NewInlineMenu(telejoon.
			NewStaticText("Custom Inline Menu"), telejoon.NewDeferredInlineActionBuilder(deferredBuilder))
}
