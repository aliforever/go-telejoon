package telejoon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"golang.org/x/text/language"
)

// === Payload transaction edge cases (second review round) ===

func TestSameStateReEntryKeepsPayload(t *testing.T) {
	stateCart := NewState[purchaseData]("Cart")

	repo := NewMockUserRepository()
	repo.SetState(1, "Cart")

	dataRepo := NewDefaultStateDataRepository()

	var received purchaseData

	engine := New(repo, stateWelcome).WithStateDataRepository(dataRepo)
	engine.Add(
		Menu(stateWelcome, S("welcome")),
		Menu(stateCart, S("cart")).
			Buttons(GoTo(S("back"), stateWelcome)).
			OnText(func(ctx *Ctx, data *purchaseData, text string) Action {
				received = *data

				// The quantity-increment idiom: re-enter the same state with
				// an updated payload.
				return ctx.GoToWith(stateCart, purchaseData{ProductID: data.ProductID, Qty: data.Qty + 1})
			}),
	)

	// Seed the payload as if a buy button had brought the user here.
	if err := dataRepo.SetUserStateData(1, "Cart", []byte(`{"ProductID":7,"Qty":1}`)); err != nil {
		t.Fatal(err)
	}

	_, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "➕").Build())

	if received.ProductID != 7 || received.Qty != 1 {
		t.Fatalf("first entry received %+v, want the seeded payload", received)
	}

	// The same-state re-entry must have stored the incremented payload, not
	// deleted it.
	raw, err := dataRepo.GetUserStateData(1, "Cart")
	if err != nil || !strings.Contains(string(raw), `"Qty":2`) {
		t.Fatalf("payload after same-state re-entry = %s, %v; want Qty 2", raw, err)
	}
}

func TestCrossStatePayloadlessEntryClearsStaleTarget(t *testing.T) {
	stateTyped := NewState[purchaseData]("StaleTyped")

	repo := NewMockUserRepository()
	repo.SetState(1, "Welcome")

	dataRepo := NewDefaultStateDataRepository()

	// An orphaned payload, as a failed transition could have left behind.
	if err := dataRepo.SetUserStateData(1, "StaleTyped", []byte(`{"ProductID":99,"Qty":9}`)); err != nil {
		t.Fatal(err)
	}

	var received purchaseData

	menuRef := NewInlineMenuRef("Any")

	engine := New(repo, stateWelcome).WithStateDataRepository(dataRepo)
	engine.Add(
		Menu(stateWelcome, S("welcome")),
		Menu(stateTyped, S("typed")).
			OnText(func(ctx *Ctx, data *purchaseData, text string) Action {
				received = *data

				return Stop()
			}),
		InlineMenuFor(menuRef, S("any")),
	)

	_, bot := newTelegramServer(t)

	// Payload-less transitions into a typed state are compile-prevented, so
	// the only way in without a payload is a (possibly forged) internal @s
	// callback — client-controlled, so the engine must defend itself.
	engine.Process(context.Background(), bot,
		NewTestUpdate().WithCallbackQuery(1, "Any:@s:StaleTyped").Build())

	if raw, _ := dataRepo.GetUserStateData(1, "StaleTyped"); len(raw) != 0 {
		t.Fatalf("stale target payload survived the entry: %s", raw)
	}

	// ...and the next update must see a zero payload, never the stale one.
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	if received.ProductID != 0 {
		t.Fatalf("handler received stale payload %+v, want zero", received)
	}
}

// setStateFailingRepo fails only state publication.
type setStateFailingRepo struct{ MockUserRepository }

func (r *setStateFailingRepo) SetUserState(id int64, state string) error {
	return fmt.Errorf("database is down")
}

func TestSwitchStateRollsBackPayloadWhenPublishFails(t *testing.T) {
	stateBuy := NewState[purchaseData]("RollbackBuy")

	repo := &setStateFailingRepo{}
	repo.SetState(1, "Welcome")

	dataRepo := NewDefaultStateDataRepository()

	logger, hook := logrustest.NewNullLogger()

	engine := New(repo, stateWelcome, NewOptions().SetLogger(logger)).
		WithStateDataRepository(dataRepo)
	engine.Add(
		Menu(stateWelcome, S("welcome")).
			Buttons(GoToWith(S("buy"), stateBuy, purchaseData{ProductID: 5})),
		Menu(stateBuy, S("buy")),
	)

	_, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "buy").Build())

	// The payload stored for the target must be rolled back when publishing
	// the state fails.
	if raw, _ := dataRepo.GetUserStateData(1, "RollbackBuy"); len(raw) != 0 {
		t.Fatalf("orphaned payload after failed publish: %s", raw)
	}

	if !strings.Contains(lastLogEntry(hook), "error_setting_user_state") {
		t.Fatalf("expected error_setting_user_state report, got %q", lastLogEntry(hook))
	}
}

func TestSwitchUserStateDeletesLeftStatePayload(t *testing.T) {
	stateTyped := NewState[purchaseData]("ExternalTyped")
	stateOther := NewState[NoData]("ExternalOther")

	repo := NewMockUserRepository()
	repo.SetState(1, "ExternalTyped")

	dataRepo := NewDefaultStateDataRepository()

	if err := dataRepo.SetUserStateData(1, "ExternalTyped", []byte(`{"ProductID":3,"Qty":1}`)); err != nil {
		t.Fatal(err)
	}

	engine := New(repo, stateWelcome).WithStateDataRepository(dataRepo)
	engine.Add(
		Menu(stateWelcome, S("welcome")),
		Menu(stateTyped, S("typed")),
		Menu(stateOther, S("other")),
	)

	_, bot := newTelegramServer(t)

	if err := engine.SwitchUserState(context.Background(), bot, 1, stateOther); err != nil {
		t.Fatalf("SwitchUserState: %v", err)
	}

	if raw, _ := dataRepo.GetUserStateData(1, "ExternalTyped"); len(raw) != 0 {
		t.Fatalf("payload of the externally-left state survived: %s", raw)
	}

	if !repo.AssertStateSet(1, "ExternalOther") {
		t.Fatal("external transition did not publish the new state")
	}
}

// === Middleware contract edge cases ===

func TestSelfRedirectingMiddlewareRunsOnce(t *testing.T) {
	stateLoop := NewState[NoData]("Loop")

	repo := NewMockUserRepository()
	repo.SetState(1, "Loop")

	runs := 0

	logger, hook := logrustest.NewNullLogger()

	engine := New(repo, stateWelcome, NewOptions().SetLogger(logger))
	engine.Add(
		Menu(stateWelcome, S("welcome")),
		Menu(stateLoop, S("loop")).
			Use(func(ctx *Ctx) Action {
				runs++

				return ctx.GoTo(stateLoop) // redirect into itself
			}),
	)

	ts, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	// Marked on entry, the re-entered chain is skipped: one run, one render,
	// no depth-exceeded error.
	if runs != 1 {
		t.Fatalf("self-redirecting middleware runs = %d, want 1", runs)
	}

	if sends := ts.callsOf("sendMessage"); len(sends) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(sends))
	}

	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, "state_switch_depth_exceeded") {
			t.Fatal("self-redirect exhausted the switch depth; want the chain skipped on re-entry")
		}
	}
}

// === Chooser exact-language lookups ===

func TestChooserFallsBackToTagWhenTranslationMissing(t *testing.T) {
	dir := t.TempDir()

	// Only English translates the chooser strings; Persian does not.
	if err := os.WriteFile(filepath.Join(dir, "pick.en.toml"), []byte(
		"[Pick]\nText = \"Pick your language\"\nButton = \"English\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "pick.fa.toml"), []byte(
		"Greeting = \"Salam\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	languages, err := NewLanguageBuilder(language.English).
		AddTOML(filepath.Join(dir, "pick.en.toml"), filepath.Join(dir, "pick.fa.toml")).
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	statePick := NewState[NoData]("Pick")

	langRepo := NewMockUserLanguageRepository()
	langRepo.SetLanguage(1, "en")

	repo := NewMockUserRepository()
	repo.SetState(1, "Pick")

	engine := New(repo, statePick).
		WithLanguageConfig(NewLanguageConfig(languages, langRepo).
			WithChangeLanguageMenu(statePick, false))

	// Registration must NOT panic with a false duplicate-label error: the
	// Persian label falls back to its tag, not to the English translation.
	ts, bot := newTelegramServer(t)

	// Render the chooser: unmatched text falls through to the render step.
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	sends := ts.callsOf("sendMessage")
	if len(sends) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1 (chooser)", len(sends))
	}

	if !strings.Contains(sends[0].body, "English") || !strings.Contains(sends[0].body, `"fa"`) {
		t.Fatalf("chooser keyboard = %s, want the English label and the fa tag", sends[0].body)
	}

	// The chooser text must not duplicate the English line for Persian.
	if strings.Count(sends[0].body, "Pick your language") != 1 {
		t.Fatalf("chooser text = %s, want the English line exactly once", sends[0].body)
	}

	// Pressing the tag-fallback label selects Persian.
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "fa").Build())

	if len(langRepo.SetLanguageCalls) == 0 || langRepo.SetLanguageCalls[0].Language != "fa" {
		t.Fatalf("language calls = %+v, want fa selected", langRepo.SetLanguageCalls)
	}
}

// === Message key validation ===

func TestMsgPKeysAreValidated(t *testing.T) {
	resetMsgRegistryForTest()
	t.Cleanup(resetMsgRegistryForTest)

	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	type greetParams struct {
		Name string `json:"Name"`
	}

	NewMsg("Greeting")                  // exists
	NewMsgP[greetParams]("WelcomeUser") // exists, needs params
	NewMsgP[greetParams]("Nope.Missing")
	NewMsg("Also.Missing")

	missing := languages.validateMsgs()
	if len(missing) != 2 || missing[0] != "Also.Missing" || missing[1] != "Nope.Missing" {
		t.Fatalf("validateMsgs = %v, want [Also.Missing Nope.Missing]", missing)
	}
}

// === Forced chooser vs. middleware-assigned language ===

func TestForcedChooserSkippedWhenMiddlewareSetsLanguage(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	chooseLanguage := NewState[NoData]("PickLang")

	repo := NewMockUserRepository() // brand-new user: no state, no language

	engine := New(repo, stateWelcome).
		WithLanguageConfig(NewLanguageConfig(languages, NewMockUserLanguageRepository()).
			WithChangeLanguageMenu(chooseLanguage, true)).
		Use(func(ctx *Ctx) Action {
			// An auth middleware that also defaults the language.
			if ctx.Language() == nil {
				if err := ctx.ChangeLanguage("en"); err != nil {
					return Error(err)
				}
			}

			return Next()
		})

	welcomed := false

	engine.Add(
		Menu(stateWelcome, S("welcome")).
			OnText(func(ctx *Ctx, _ *NoData, text string) Action {
				welcomed = true

				return Stop()
			}),
	)

	_, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	if !welcomed {
		t.Fatal("update was not dispatched to the default state")
	}

	if repo.AssertStateSet(1, "PickLang") {
		t.Fatal("forced chooser fired although the middleware assigned a language")
	}
}

// === Shutdown timeout ===

// blockingProcessor accepts every update and blocks until released.
type blockingProcessor struct {
	started chan struct{}
	release chan struct{}
}

func (p *blockingProcessor) canProcess(update tgbotapi.Update) bool { return true }

func (p *blockingProcessor) Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	close(p.started)
	<-p.release
}

// newPollingServer returns a bot client backed by a test server that serves
// one update on the first getUpdates call and empty batches afterwards.
func newPollingServer(t *testing.T, served *atomic.Bool) *tgbotapi.Bot {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch path.Base(r.URL.Path) {
		case "getUpdates":
			if served != nil && !served.Swap(true) {
				fmt.Fprint(w, `{"ok":true,"result":[{"update_id":1,"message":{"message_id":1,"date":1,`+
					`"chat":{"id":1,"type":"private"},"from":{"id":1},"text":"hi"}}]}`)

				return
			}

			fmt.Fprint(w, `{"ok":true,"result":[]}`)
		case "sendMessage":
			fmt.Fprint(w, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"},"text":"ok"}}`)
		default:
			fmt.Fprint(w, `{"ok":true,"result":true}`)
		}
	}))

	t.Cleanup(server.Close)

	return tgbotapi.New("test-token", tgbotapi.WithAPIURL(server.URL))
}

func TestStartWithShutdownTimeout(t *testing.T) {
	var served atomic.Bool

	bot := newPollingServer(t, &served)

	processor := &blockingProcessor{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}

	t.Cleanup(func() { close(processor.release) })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-processor.started
		cancel()
	}()

	err := StartWithShutdownTimeout(ctx, bot, processor, 50*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "shutdown timed out") {
		t.Fatalf("Start err = %v, want a shutdown timeout", err)
	}
}
