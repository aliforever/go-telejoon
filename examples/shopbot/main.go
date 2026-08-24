package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
	"github.com/aliforever/go-telejoon"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

// botEngine is kept package-level so admin commands can demonstrate the
// external engine operations (SwitchUserState, SendInlineMenu).
var botEngine *telejoon.Engine

func main() {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		fmt.Println("usage: BOT_TOKEN=... go run ./examples/shopbot")
		fmt.Println("optional: ERROR_GROUP_ID=<telegram group id for error reports>")

		return
	}

	// v2's New performs no network I/O; validate the token explicitly with Me.
	client := tgbotapi.New(token)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if _, err := client.Me(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "create bot:", err)

		return
	}

	logger := logrus.New()

	// === Languages (en + fa/RTL) ===

	languages, err := telejoon.NewLanguageBuilder(language.English, language.Persian).
		RegisterTomlFormat(localePaths()).
		Build()
	if err != nil {
		logger.Fatal("build languages: ", err)
	}

	languageConfig := telejoon.NewLanguageConfig(languages, telejoon.NewDefaultUserLanguageRepository()).
		// Force-choose: brand-new users pick a language before anything else,
		// then land on the default state. The menu is auto-registered.
		WithChangeLanguageMenu("ChooseLanguage", true).
		WithReverseButtonOrderInRowForRTL()

	// === Engine ===

	options := telejoon.NewOptions().SetLogger(logger)
	if groupID, _ := strconv.ParseInt(os.Getenv("ERROR_GROUP_ID"), 10, 64); groupID != 0 {
		options.SetErrorGroupID(groupID)
	}

	botEngine = telejoon.New(telejoon.NewDefaultUserRepository(), stateWelcome, options).
		WithLanguageConfig(languageConfig).
		// Required for GoToWith payloads to survive to the user's NEXT
		// message (e.g. the checkout address). See store.go.
		WithStateDataRepository(&memoryStateDataRepository{}).
		WithPanicHandler(func(client *tgbotapi.Bot, update tgbotapi.Update, err any, trace string) {
			// Panics are recovered per update; the bot stays up.
			logger.Errorf("panic: %v\n%s", err, trace)
		}).
		// Engine middlewares run for every private update, in order,
		// BEFORE menu middlewares. Two caveats: the forced first-time
		// language selection redirects before engine middlewares run, and
		// state payloads are not loaded yet at this point — don't read
		// StateData here.
		Use(requestLogger).
		Use(cancelCommand)

	// Engine-global route: buttons bound to it work from any inline menu.
	trackRoute = botEngine.Route("track", func(ctx *telejoon.Ctx, orderID int64) telejoon.Action {
		return ctx.AnswerCallback(fmt.Sprintf("order #%d: in transit (global route)", orderID), false)
	})

	botEngine.Add(
		welcomeMenu(),
		productsMenu(),
		checkoutMenu(),
		supportMenu(),
		profileMenu(),
		adminMenu(),
		broadcastMenu(),
		orderReadyMenu(),
		catalogMenu(),
		detailsMenu(),
		aboutMenu(),
	)

	// Validate catches broken static cross-references (GoTo/ShowInline/
	// OpenMenu/StateBtn targets) at startup. ButtonsFunc output is
	// per-request and cannot be validated here.
	if err := botEngine.Validate(); err != nil {
		logger.Fatal("validate: ", err)
	}

	// === Group and channel processors ===
	//
	// Group/channel handlers have a different shape: return true to continue
	// to the next handler, false to stop.

	group := telejoon.NewGroupHandlers().
		AddMiddleware(func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool {
			logger.Debug("group update")

			return true
		}).
		AddHandler(func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool {
			if update.Message != nil && strings.Contains(update.Message.Text, "/shop") {
				_, err := client.Message().
					ChatID(update.Message.Chat.Id).
					Text("Open a private chat with me to browse the shop!").
					Send(ctx)
				if err != nil {
					logger.Error("group reply: ", err)
				}

				return false
			}

			return true
		})

	channel := telejoon.NewChannelHandlers().
		AddHandler(func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool {
			if post := update.ChannelPost; post != nil {
				logger.Infof("channel post: %s", post.Text)
			}

			return true
		})

	// MultiProcessor routes each update to the first processor whose
	// canProcess matches (private, then group, then channel).
	multi := telejoon.NewMultiProcessor(botEngine, group, channel)

	// === Start ===

	logger.Info("shopbot running — Ctrl+C for graceful shutdown")

	// Start long-polls and blocks until the context is cancelled, waiting for
	// in-flight updates to finish.
	if err := telejoon.Start(ctx, client, multi); err != nil {
		logger.Info("stopped: ", err)
	}
}

// requestLogger demonstrates an engine middleware that stamps per-request
// session values (read later by K/F texts) and logs the update type.
func requestLogger(ctx *telejoon.Ctx) telejoon.Action {
	ctx.Set(nameKey, ctx.FirstName())
	ctx.Set(requestTagKey, "req")
	ctx.Set(requestTagKey2, 1) // same key name, different type — no collision

	return telejoon.Next()
}

// cancelCommand is a global /cancel: works in every state because it is an
// engine middleware returning a transition action instead of Next.
func cancelCommand(ctx *telejoon.Ctx) telejoon.Action {
	if ctx.IsCommand() && ctx.Command().Name == "cancel" {
		return ctx.GoTo(stateWelcome)
	}

	return telejoon.Next()
}

// localePaths resolves the locale files whether the bot runs from the
// repository root (go run ./examples/shopbot) or its own directory.
func localePaths() []string {
	if _, err := os.Stat("locale.en.toml"); err == nil {
		return []string{"locale.en.toml", "locale.fa.toml"}
	}

	return []string{"examples/shopbot/locale.en.toml", "examples/shopbot/locale.fa.toml"}
}
