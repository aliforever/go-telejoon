package telejoon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"golang.org/x/text/language"
)

// === Recording Telegram test server ===
//
// The v2 client's WithAPIURL lets the full Engine.Process pipeline run
// against a local server that records every API call.

type recordedCall struct {
	method string
	body   string
}

type telegramServer struct {
	mu     sync.Mutex
	calls  []recordedCall
	server *httptest.Server
}

func newTelegramServer(t *testing.T) (*telegramServer, *tgbotapi.Bot) {
	t.Helper()

	ts := &telegramServer{}

	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)

		ts.mu.Lock()
		ts.calls = append(ts.calls, recordedCall{method: path.Base(r.URL.Path), body: string(raw)})
		ts.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		if path.Base(r.URL.Path) == "sendMessage" {
			fmt.Fprint(w, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"},"text":"ok"}}`)

			return
		}

		fmt.Fprint(w, `{"ok":true,"result":true}`)
	}))

	t.Cleanup(ts.server.Close)

	return ts, tgbotapi.New("test-token", tgbotapi.WithAPIURL(ts.server.URL))
}

func (ts *telegramServer) callsOf(method string) []recordedCall {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	var out []recordedCall

	for _, call := range ts.calls {
		if call.method == method {
			out = append(out, call)
		}
	}

	return out
}

// waitFor polls cond until it holds or the deadline passes — API answers
// fired from goroutines (answerCallback) are not synchronized with Process.
func (ts *telegramServer) waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

// === Middleware semantics ===

func TestMiddlewareCountsAcrossStateSwitch(t *testing.T) {
	stateA := NewState[NoData]("A")
	stateB := NewState[NoData]("B")

	repo := NewMockUserRepository()
	repo.SetState(1, "A")

	var global, mwA, mwB int

	engine := New(repo, stateA).
		Use(func(ctx *Ctx) Action { global++; return Next() })

	engine.Add(
		Menu(stateA, S("a")).
			Use(func(ctx *Ctx) Action { mwA++; return Next() }).
			Buttons(GoTo(S("to B"), stateB)),
		Menu(stateB, S("b")).
			Use(func(ctx *Ctx) Action { mwB++; return Next() }),
	)

	ts, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "to B").Build())

	// Documented semantics: the global chain runs once per update; the source
	// menu's chain runs to dispatch the update; the destination menu's chain
	// runs because the menu is entered to be rendered.
	if global != 1 || mwA != 1 || mwB != 1 {
		t.Fatalf("middleware counts = global:%d A:%d B:%d, want 1/1/1", global, mwA, mwB)
	}

	if !repo.AssertStateSet(1, "B") {
		t.Fatal("state was not switched to B")
	}

	// Only the destination menu renders a message.
	if sends := ts.callsOf("sendMessage"); len(sends) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(sends))
	}
}

func TestInlineMenuMiddlewareRunsOnceOnRefresh(t *testing.T) {
	stateMain := NewState[NoData]("Main")
	ref := NewInlineMenuRef("Catalog")

	var mw int

	menu := InlineMenuFor(ref, S("catalog"))
	menu.Use(func(ctx *Ctx) Action { mw++; return Next() })

	refresh := menu.Route("refresh", func(ctx *Ctx, _ NoData) Action {
		return Edit(ref)
	})
	menu.Buttons(Do(S("refresh"), refresh, NoData{}))

	engine := New(NewMockUserRepository(), stateMain)
	engine.Add(Menu(stateMain, S("main")), menu)

	ts, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot,
		NewTestUpdate().WithCallbackQuery(1, "Catalog:refresh:").Build())

	// The chain guards the callback dispatch; the Edit re-render must not
	// re-run it.
	if mw != 1 {
		t.Fatalf("inline middleware runs = %d, want 1", mw)
	}

	if edits := ts.callsOf("editMessageText"); len(edits) != 1 {
		t.Fatalf("editMessageText calls = %d, want 1", len(edits))
	}

	// The route did not answer the callback; the engine must.
	ts.waitFor(t, "auto answerCallbackQuery", func() bool {
		return len(ts.callsOf("answerCallbackQuery")) == 1
	})
}

func TestOnceMiddlewareRunsOnceAcrossMenus(t *testing.T) {
	stateA := NewState[NoData]("OA")
	stateB := NewState[NoData]("OB")

	repo := NewMockUserRepository()
	repo.SetState(1, "OA")

	runs := 0
	shared := Once(func(ctx *Ctx) Action { runs++; return Next() })

	engine := New(repo, stateA)
	engine.Add(
		Menu(stateA, S("a")).Use(shared).Buttons(GoTo(S("to B"), stateB)),
		Menu(stateB, S("b")).Use(shared),
	)

	_, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "to B").Build())

	if runs != 1 {
		t.Fatalf("Once-wrapped middleware runs = %d, want 1", runs)
	}
}

func TestDispatchOnlySkipsSwitchRenderPass(t *testing.T) {
	stateA := NewState[NoData]("DA")
	stateB := NewState[NoData]("DB")

	repo := NewMockUserRepository()
	repo.SetState(1, "DA")

	runs := 0

	engine := New(repo, stateA)
	engine.Add(
		Menu(stateA, S("a")).Buttons(GoTo(S("to B"), stateB)),
		Menu(stateB, S("b")).Use(DispatchOnly(func(ctx *Ctx) Action { runs++; return Next() })),
	)

	_, bot := newTelegramServer(t)

	// The switch renders B but does not dispatch to it: skipped.
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "to B").Build())

	if runs != 0 {
		t.Fatalf("DispatchOnly middleware runs on switch-render = %d, want 0", runs)
	}

	// A fresh update dispatched to B runs it.
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hello").Build())

	if runs != 1 {
		t.Fatalf("DispatchOnly middleware runs on dispatch = %d, want 1", runs)
	}
}

// === State payload lifecycle ===

type purchaseData struct {
	ProductID int64
	Qty       int
}

func TestStatePayloadLifecycle(t *testing.T) {
	stateShop := NewState[NoData]("Shop")
	stateBuy := NewState[purchaseData]("Buy")

	repo := NewMockUserRepository()
	repo.SetState(1, "Shop")

	dataRepo := NewDefaultStateDataRepository()

	var received purchaseData

	engine := New(repo, stateShop).WithStateDataRepository(dataRepo)
	engine.Add(
		Menu(stateShop, S("shop")).
			Buttons(GoToWith(S("buy"), stateBuy, purchaseData{ProductID: 42})),
		Menu(stateBuy, S("buying")).
			OnText(func(ctx *Ctx, data *purchaseData, text string) Action {
				received = *data

				return ctx.GoTo(stateShop)
			}),
	)

	_, bot := newTelegramServer(t)

	// The transition carries the payload...
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "buy").Build())

	// ...and the NEXT update reloads it from the repository.
	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "my address").Build())

	if received.ProductID != 42 {
		t.Fatalf("handler received payload %+v, want ProductID 42", received)
	}

	// Leaving the state deleted its payload: it can never resurface stale.
	raw, err := dataRepo.GetUserStateData(1, "Buy")
	if err != nil || len(raw) != 0 {
		t.Fatalf("payload after leaving = %s, %v; want deleted", raw, err)
	}
}

type failingStateDataRepo struct{}

func (failingStateDataRepo) SetUserStateData(userID int64, state string, data []byte) error {
	return nil
}

func (failingStateDataRepo) GetUserStateData(userID int64, state string) ([]byte, error) {
	return nil, fmt.Errorf("database is down")
}

func (failingStateDataRepo) DeleteUserStateData(userID int64, state string) error { return nil }

func TestStateDataLoadErrorAbortsProcessing(t *testing.T) {
	stateHome := NewState[NoData]("Home")
	stateTyped := NewState[purchaseData]("Typed")

	repo := NewMockUserRepository()
	repo.SetState(1, "Typed")

	logger, hook := logrustest.NewNullLogger()

	engine := New(repo, stateHome, NewOptions().SetLogger(logger)).
		WithStateDataRepository(failingStateDataRepo{})

	handlerCalled := false

	engine.Add(
		Menu(stateHome, S("home")),
		Menu(stateTyped, S("typed")).
			OnText(func(ctx *Ctx, data *purchaseData, text string) Action {
				handlerCalled = true

				return Stop()
			}),
	)

	_, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hello").Build())

	if handlerCalled {
		t.Fatal("handler ran with an unloadable payload; want the update aborted")
	}

	if !strings.Contains(lastLogEntry(hook), "state_data_load") {
		t.Fatalf("expected a state_data_load error report, got %q", lastLogEntry(hook))
	}
}

func lastLogEntry(hook *logrustest.Hook) string {
	entries := hook.AllEntries()
	if len(entries) == 0 {
		return ""
	}

	return entries[len(entries)-1].Message
}

// === Keyboard rendering ===

func TestKeyboardReRenderedAfterHandlerFallThrough(t *testing.T) {
	stateCounter := NewState[NoData]("Counter")

	repo := NewMockUserRepository()
	repo.SetState(1, "Counter")

	count := 0

	engine := New(repo, stateCounter)
	engine.Add(
		Menu(stateCounter, S("counter")).
			ButtonsFunc(func(ctx *Ctx, _ *NoData) []*Button {
				return []*Button{Raw(S(fmt.Sprintf("n=%d", count)))}
			}).
			OnText(func(ctx *Ctx, _ *NoData, text string) Action {
				count++

				return Next() // fall through to the render step
			}),
	)

	ts, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "tick").Build())

	sends := ts.callsOf("sendMessage")
	if len(sends) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(sends))
	}

	// The handler mutated what ButtonsFunc renders; the sent keyboard must
	// reflect the mutation, not the pre-dispatch render.
	if !strings.Contains(sends[0].body, "n=1") {
		t.Fatalf("sent keyboard = %s, want the post-handler re-render (n=1)", sends[0].body)
	}
}

func TestKeyboardRemovedWhenNoButtonsVisible(t *testing.T) {
	statePlain := NewState[NoData]("Plain")

	repo := NewMockUserRepository()
	repo.SetState(1, "Plain")

	engine := New(repo, statePlain)
	engine.Add(Menu(statePlain, S("just text")))

	ts, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	sends := ts.callsOf("sendMessage")
	if len(sends) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1", len(sends))
	}

	// A state with no visible buttons must pull any keyboard a previous
	// state left behind.
	if !strings.Contains(sends[0].body, `"remove_keyboard":true`) {
		t.Fatalf("sendMessage body = %s, want remove_keyboard", sends[0].body)
	}
}

// === Languages ===

func TestForcedLanguageChooserRunsAfterGlobalMiddleware(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	chooseLanguage := NewState[NoData]("PickLanguage")

	repo := NewMockUserRepository() // brand-new user: no state, no language

	globalRuns := 0

	engine := New(repo, stateWelcome).
		WithLanguageConfig(NewLanguageConfig(languages, NewMockUserLanguageRepository()).
			WithChangeLanguageMenu(chooseLanguage, true)).
		Use(func(ctx *Ctx) Action { globalRuns++; return Next() })

	engine.Add(Menu(stateWelcome, S("welcome")))

	ts, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	// Auth-style global middlewares must not be bypassed by first contact.
	if globalRuns != 1 {
		t.Fatalf("global middleware runs = %d, want 1", globalRuns)
	}

	if !repo.AssertStateSet(1, "PickLanguage") {
		t.Fatal("new user was not redirected to the language chooser")
	}

	sends := ts.callsOf("sendMessage")
	if len(sends) != 1 {
		t.Fatalf("sendMessage calls = %d, want 1 (the chooser)", len(sends))
	}

	// No chooser translations in the test locales: labels fall back to tags.
	if !strings.Contains(sends[0].body, `"en"`) || !strings.Contains(sends[0].body, `"fa"`) {
		t.Fatalf("chooser body = %s, want one button per language tag", sends[0].body)
	}
}

func TestUnknownStoredLanguageFallsBackToDefault(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	langRepo := NewMockUserLanguageRepository()
	langRepo.SetLanguage(1, "de") // not among the configured languages

	repo := NewMockUserRepository()
	repo.SetState(1, "Welcome")

	var seenTag string

	engine := New(repo, stateWelcome).
		WithLanguageConfig(NewLanguageConfig(languages, langRepo))

	engine.Add(
		Menu(stateWelcome, S("welcome")).
			OnText(func(ctx *Ctx, _ *NoData, text string) Action {
				if lang := ctx.Language(); lang != nil {
					seenTag = lang.Tag()
				}

				return Stop()
			}),
	)

	_, bot := newTelegramServer(t)

	engine.Process(context.Background(), bot, NewTestUpdate().WithMessage(1, 1, "hi").Build())

	if seenTag != "en" {
		t.Fatalf("language for unknown stored tag = %q, want the default %q", seenTag, "en")
	}
}

func TestMsgPWithTypedParams(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	type greetParams struct {
		Name string `json:"Name"`
	}

	greet := NewMsgP[greetParams]("WelcomeUser")

	// fa has no WelcomeUser translation: the default language fills in.
	ctx := NewTestUpdate().WithMessage(1, 1, "hi").BuildCtxWithLanguage("Welcome", languages.GetByTag("fa"))

	if got := greet.T(greetParams{Name: "Bob"})(ctx); got != "Hello Bob" {
		t.Fatalf("MsgP.T = %q, want %q", got, "Hello Bob")
	}
}

func TestValidateChecksDeclaredMessageKeys(t *testing.T) {
	resetMsgRegistryForTest()
	t.Cleanup(resetMsgRegistryForTest)

	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	existing := NewMsg("Greeting")

	engine := New(NewMockUserRepository(), stateWelcome).
		WithLanguageConfig(NewLanguageConfig(languages, NewMockUserLanguageRepository()))
	engine.Add(Menu(stateWelcome, existing.T()))

	if err := engine.Validate(); err != nil {
		t.Fatalf("Validate with an existing key: %v", err)
	}

	// A key no locale defines must fail Validate.
	NewMsg("No.Such.Key")

	if err := engine.Validate(); err == nil || !strings.Contains(err.Error(), "No.Such.Key") {
		t.Fatalf("Validate with a missing key = %v, want messages_not_translated", err)
	}
}

func TestLanguageBuilderRejectsDuplicatesAndMissingDefault(t *testing.T) {
	if _, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.en.toml").
		Build(); err == nil || !strings.Contains(err.Error(), "duplicate_language_file") {
		t.Fatalf("duplicate file err = %v, want duplicate_language_file", err)
	}

	if _, err := NewLanguageBuilder(language.German).
		AddTOML("testdata/locale.en.toml").
		Build(); err == nil || !strings.Contains(err.Error(), "no_message_file_for_default_language") {
		t.Fatalf("missing default err = %v, want no_message_file_for_default_language", err)
	}
}

func TestRTLAutoDetection(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		AddTOML("testdata/locale.en.toml", "testdata/locale.fa.toml").
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	if languages.GetByTag("fa").IsRTL() != true {
		t.Fatal("Persian must be auto-detected as RTL")
	}

	if languages.GetByTag("en").IsRTL() != false {
		t.Fatal("English must not be RTL")
	}
}

// === Per-chat serialization ===

func TestChatSerializerFIFO(t *testing.T) {
	serializer := &chatSerializer{tails: map[int64]chan struct{}{}}

	var mu sync.Mutex

	var order []int

	var wg sync.WaitGroup

	const updates = 50

	for i := 0; i < updates; i++ {
		wait, link := serializer.chain(7)

		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			defer serializer.finish(7, link)

			if wait != nil {
				<-wait
			}

			mu.Lock()
			order = append(order, i)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	for i := 0; i < updates; i++ {
		if order[i] != i {
			t.Fatalf("execution order = %v, want arrival order", order)
		}
	}

	if len(serializer.tails) != 0 {
		t.Fatalf("serializer tails leaked: %d entries", len(serializer.tails))
	}
}
