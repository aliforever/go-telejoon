//go:build manual

// This file is a manual playground, not an automated test: it requires a real
// BOT_TOKEN, locale files on the author's machine, and blocks on live updates.
// Run it explicitly with: go test -tags manual -run TestStart
package telejoon_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
	"github.com/aliforever/go-telejoon"
	"golang.org/x/text/language"
)

var (
	stateWelcome = telejoon.NewState[telejoon.NoData]("Welcome")
	stateInfo    = telejoon.NewState[telejoon.NoData]("Info")

	stateChangeLanguage = telejoon.NewState[telejoon.NoData]("ChangeLanguage")

	menuInfo   = telejoon.NewInlineMenuRef("Info")
	menuInfo2  = telejoon.NewInlineMenuRef("Info2")
	menuCustom = telejoon.NewInlineMenuRef("CustomInline")

	nameKey = telejoon.NewKey[string]("name")
)

type callbackArgs struct {
	Data string
}

func TestStart(t *testing.T) {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		t.Skip("BOT_TOKEN is not set")
	}

	// v2's New performs no network I/O; validate explicitly with Me.
	client := tgbotapi.New(botToken)

	ctx := context.Background()

	if _, err := client.Me(ctx); err != nil {
		t.Fatal(err)
	}

	languages, err := telejoon.NewLanguageBuilder(language.English).
		AddTOML(
			`C:\golang\src\github.com\aliforever\go-telejoon\locale.en.toml`,
			`C:\golang\src\github.com\aliforever\go-telejoon\locale.fa.toml`,
		).Build()
	if err != nil {
		t.Fatal(err)
	}

	languageConfig := telejoon.NewLanguageConfig(languages, telejoon.NewDefaultUserLanguageRepository()).
		WithChangeLanguageMenu(stateChangeLanguage, true)

	info := telejoon.InlineMenuFor(menuInfo, telejoon.S("Info Inline Menu"))

	callback1 := info.Route("callback_1", func(ctx *telejoon.Ctx, args callbackArgs) telejoon.Action {
		return ctx.AnswerCallback(fmt.Sprintf("Callback 1 Clicked with args: %s", args.Data), false)
	})

	info.Buttons(
		telejoon.URL(telejoon.S("Google"), telejoon.S("https://google.com")),
		telejoon.Alert(telejoon.L("Info.Hello"), "HI!"),
		telejoon.Alert(telejoon.S("Hello"), "Hello Friend"),
		telejoon.Alert(telejoon.S("Hello 2"), "Hello Friend 2"),
		telejoon.Alert(telejoon.S("Hello 3"), "Hello Friend 3"),
		telejoon.Do(telejoon.S("Callback 1"), callback1, callbackArgs{Data: "data"}),
		telejoon.Do(telejoon.S("Callback 2"), callback1, callbackArgs{Data: "data2"}),
		telejoon.OpenMenuEdit(telejoon.S("Change Menu to Info 2"), menuInfo2),
	).MaxPerRow(3)

	engine := telejoon.New(telejoon.NewDefaultUserRepository(), stateWelcome,
		telejoon.NewOptions().SetErrorGroupID(81997375)).
		WithLanguageConfig(languageConfig).
		WithPanicHandler(func(client *tgbotapi.Bot, update tgbotapi.Update, err any, stack string) {
			fmt.Println("Panic Handler", update, "\n", stack)
		}).
		Use(func(ctx *telejoon.Ctx) telejoon.Action {
			if ctx.Update.Message != nil && ctx.Update.Message.Text == "panic" {
				panic("Panic Test")
			}
			return telejoon.Next()
		}).
		Use(func(ctx *telejoon.Ctx) telejoon.Action {
			fmt.Println("update inside middleware", ctx.Update)
			return telejoon.Next()
		}).
		Add(
			telejoon.Menu(stateWelcome, telejoon.L("Welcome.Main")).
				Buttons(
					telejoon.GoTo(telejoon.L("Welcome.ChangeLanguageBtn"), stateChangeLanguage),
					telejoon.Reply(telejoon.S("Hello"), telejoon.S("You said Hello")),
					telejoon.GoTo(telejoon.S("Info State"), stateInfo),
					telejoon.ShowInline(telejoon.S("Info"), menuInfo),
				).
				Use(func(ctx *telejoon.Ctx) telejoon.Action {
					ctx.Set(nameKey, "Ali")
					return telejoon.Next()
				}).
				OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
					if text == "Hello Bro" {
						fmt.Println("changing name to:", "Ali 2")
						ctx.Set(nameKey, "Ali 2")

						return ctx.ReplyText("Hello Bro!")
					}

					return telejoon.Next()
				}),
			telejoon.Menu(stateInfo, telejoon.L("Info.Hello")).
				Buttons(telejoon.GoTo(telejoon.L("Global.Back"), stateWelcome)),
			info,
			telejoon.InlineMenuFor(menuInfo2, telejoon.S("Info2 Inline Menu")).
				Buttons(
					telejoon.AlertDialog(telejoon.S("Hello"), "Hello Friend"),
					telejoon.OpenMenuEdit(telejoon.S("CustomInline"), menuCustom),
					telejoon.OpenMenuEdit(telejoon.S("Back"), menuInfo),
				),
			telejoon.InlineMenuFor(menuCustom, telejoon.S("Custom Inline Menu")).
				ButtonsFunc(func(ctx *telejoon.Ctx) []*telejoon.InlineButton {
					return []*telejoon.InlineButton{
						telejoon.AlertDialog(telejoon.S("Hello"), "Hello Friend"),
					}
				}),
		)

	if err := engine.Validate(); err != nil {
		t.Fatal(err)
	}

	// Start long-polls and blocks until the context is cancelled.
	if err := telejoon.Start(ctx, client, engine); err != nil {
		fmt.Println("stopped:", err)
	}
}
