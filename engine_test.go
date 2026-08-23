package telejoon

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/aliforever/go-telegram-bot-api/structs"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	"golang.org/x/text/language"
)

var (
	stateWelcome  = NewState[NoData]("Welcome")
	stateCheckout = NewState[checkoutData]("Checkout")
)

type checkoutData struct {
	ProductID int64
	Qty       int
}

func testCtx(text string) *Ctx {
	return NewTestUpdate().WithMessage(1, 1, text).BuildCtx("Welcome")
}

func replyRows(t *testing.T, markup interface{}) [][]string {
	t.Helper()

	j, err := json.Marshal(markup)
	if err != nil {
		t.Fatalf("marshal markup: %v", err)
	}

	var parsed struct {
		Keyboard [][]struct {
			Text string `json:"text"`
		} `json:"keyboard"`
	}
	if err := json.Unmarshal(j, &parsed); err != nil {
		t.Fatalf("unmarshal markup %s: %v", j, err)
	}

	var rows [][]string
	for _, row := range parsed.Keyboard {
		var texts []string
		for _, btn := range row {
			texts = append(texts, btn.Text)
		}
		rows = append(rows, texts)
	}

	return rows
}

type inlineButtonRow struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
	URL          string `json:"url"`
}

func inlineRows(t *testing.T, markup interface{}) [][]inlineButtonRow {
	t.Helper()

	j, err := json.Marshal(markup)
	if err != nil {
		t.Fatalf("marshal markup: %v", err)
	}

	var parsed struct {
		InlineKeyboard [][]inlineButtonRow `json:"inline_keyboard"`
	}
	if err := json.Unmarshal(j, &parsed); err != nil {
		t.Fatalf("unmarshal markup %s: %v", j, err)
	}

	return parsed.InlineKeyboard
}

func equalRows(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

// === Typed session storage (Key[T] + generic methods) ===

func TestKeyTypedStorage(t *testing.T) {
	ctx := testCtx("hi")

	cartKey := NewKey[[]string]("cart")
	otherCartKey := NewKey[int]("cart") // same name, different type: no collision

	if _, ok := ctx.Get(cartKey); ok {
		t.Fatal("Get on empty storage should report false")
	}

	ctx.Set(cartKey, []string{"apple"})
	ctx.Set(otherCartKey, 42)

	cart, ok := ctx.Get(cartKey)
	if !ok || len(cart) != 1 || cart[0] != "apple" {
		t.Fatalf("Get(cartKey) = %v, %v", cart, ok)
	}

	if got := ctx.GetOr(otherCartKey, 0); got != 42 {
		t.Fatalf("GetOr(otherCartKey) = %d, want 42", got)
	}

	if got := ctx.GetOr(NewKey[string]("missing"), "fallback"); got != "fallback" {
		t.Fatalf("GetOr missing = %q, want fallback", got)
	}

	if !ctx.Has(cartKey) {
		t.Fatal("Has(cartKey) = false, want true")
	}

	ctx.Delete(cartKey)
	if ctx.Has(cartKey) {
		t.Fatal("Delete did not remove the value")
	}
}

// === Memoized conditions ===

func TestCondMemoizedOncePerCtx(t *testing.T) {
	adminKey := NewKey[bool]("admin")

	calls := 0
	cond := DefineCond(func(ctx *Ctx) bool {
		calls++
		return ctx.GetOr(adminKey, false)
	})

	ctx := testCtx("hi")

	if cond.eval(ctx) || cond.eval(ctx) {
		t.Fatal("condition unexpectedly true")
	}
	if calls != 1 {
		t.Fatalf("condition evaluated %d times for one ctx, want 1 (memoized)", calls)
	}

	other := testCtx("hi")
	other.Set(adminKey, true)

	if !cond.eval(other) {
		t.Fatal("condition should be true for admin ctx")
	}
	if calls != 2 {
		t.Fatalf("condition evaluated %d times, want 2 (one per ctx)", calls)
	}
	if cond.eval(ctx) {
		t.Fatal("per-request result leaked between ctxs")
	}
}

func TestKeyboardConcurrentRenderIsolation(t *testing.T) {
	adminKey := NewKey[bool]("admin")
	isAdmin := DefineCond(func(ctx *Ctx) bool { return ctx.GetOr(adminKey, false) })

	engine := New(NewMockUserRepository(), stateWelcome)
	runtime := Menu(stateWelcome, S("hi")).
		Buttons(GoTo(S("Admin"), stateWelcome).When(isAdmin)).
		compile()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			want := i%2 == 0

			ctx := testCtx("Admin")
			ctx.Set(adminKey, want)

			_, byLabel, err := engine.renderReplyKeyboard(ctx, runtime)
			if err != nil {
				t.Errorf("renderReplyKeyboard: %v", err)
				return
			}

			_, found := byLabel["Admin"]
			if found != want {
				t.Errorf("button visible=%v, want %v (cross-user condition leak)", found, want)
			}
		}(i)
	}

	wg.Wait()
}

// === Keyboard layout ===

func TestFormationNotDuplicated(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	var buttons []*Button
	for _, name := range []string{"A", "B", "C", "D", "E"} {
		buttons = append(buttons, Reply(S(name), S(name)))
	}

	runtime := Menu(stateWelcome, S("hi")).Buttons(buttons...).Formation(2, 1).compile()

	markup, _, err := engine.renderReplyKeyboard(testCtx("hi"), runtime)
	if err != nil {
		t.Fatalf("renderReplyKeyboard: %v", err)
	}

	want := [][]string{{"A", "B"}, {"C"}, {"D"}, {"E"}}
	if got := replyRows(t, markup); !equalRows(got, want) {
		t.Fatalf("formation rows = %v, want %v (base formation must not be duplicated)", got, want)
	}
}

func TestBreakMarkersProduceNoEmptyButtons(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	runtime := Menu(stateWelcome, S("hi")).
		Buttons(
			Reply(S("A"), S("A")),
			Reply(S("B"), S("B")).Alone(),
			Reply(S("C"), S("C")),
		).
		compile()

	markup, _, err := engine.renderReplyKeyboard(testCtx("hi"), runtime)
	if err != nil {
		t.Fatalf("renderReplyKeyboard: %v", err)
	}

	want := [][]string{{"A"}, {"B"}, {"C"}}
	if got := replyRows(t, markup); !equalRows(got, want) {
		t.Fatalf("rows = %v, want %v (break markers must not become empty buttons)", got, want)
	}
}

func TestChunkIntoRows(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		maxPerRow int
		formation []int
		want      [][]string
	}{
		{
			name:      "markers close rows without rendering",
			items:     []string{"", "A", "B", "", "C"},
			maxPerRow: 3,
			want:      [][]string{{"A", "B"}, {"C"}},
		},
		{
			name:  "default one per row",
			items: []string{"A", "B"},
			want:  [][]string{{"A"}, {"B"}},
		},
		{
			name:      "maxPerRow packs rows",
			items:     []string{"A", "B", "C"},
			maxPerRow: 2,
			want:      [][]string{{"A", "B"}, {"C"}},
		},
		{
			name:      "formation repeats last entry when exhausted",
			items:     []string{"A", "B", "C", "D", "E"},
			formation: []int{2, 1},
			want:      [][]string{{"A", "B"}, {"C"}, {"D"}, {"E"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := chunkIntoRows(tt.items, isEmptyString, tt.maxPerRow, tt.formation, false); !equalRows(got, tt.want) {
				t.Fatalf("chunkIntoRows = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhenUnlessVisibility(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)
	isAdmin := DefineCond(func(ctx *Ctx) bool { return false })

	runtime := Menu(stateWelcome, S("hi")).
		Buttons(
			GoTo(S("Panel"), stateWelcome).When(isAdmin),
			Reply(S("Public"), S("Public")).Unless(isAdmin),
		).
		compile()

	markup, byLabel, err := engine.renderReplyKeyboard(testCtx("Panel"), runtime)
	if err != nil {
		t.Fatalf("renderReplyKeyboard: %v", err)
	}

	if _, found := byLabel["Panel"]; found {
		t.Fatal("button with false When condition should be hidden")
	}
	if _, found := byLabel["Public"]; !found {
		t.Fatal("button with false Unless condition should be visible")
	}

	if rows := replyRows(t, markup); !equalRows(rows, [][]string{{"Public"}}) {
		t.Fatalf("rows = %v, want only the visible button", rows)
	}
}

func TestDuplicateButtonLabelIsError(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	runtime := Menu(stateWelcome, S("hi")).
		Buttons(Reply(S("Same"), S("1")), Reply(S("Same"), S("2"))).
		compile()

	if _, _, err := engine.renderReplyKeyboard(testCtx("hi"), runtime); err == nil {
		t.Fatal("expected duplicate_button_label error, got nil")
	}
}

// === Typed routes and codecs ===

type delArgs struct {
	ProductID int64
	Name      string
	InStock   bool
}

func TestPositionalCodecRoundTrip(t *testing.T) {
	codec := positionalCodec[delArgs]{}

	original := delArgs{ProductID: 42, Name: "a:b c", InStock: true}

	encoded, err := codec.Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded != original {
		t.Fatalf("round trip = %+v, want %+v", decoded, original)
	}
}

func TestPositionalCodecSingleValue(t *testing.T) {
	codec := positionalCodec[int64]{}

	encoded, err := codec.Encode(7)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded != 7 {
		t.Fatalf("round trip = %d, want 7", decoded)
	}

	if _, err := codec.Decode([]byte("1:2")); err == nil {
		t.Fatal("expected field count mismatch error, got nil")
	}
}

func TestJSONCodecRoundTrip(t *testing.T) {
	codec := JSONCodec[delArgs]()

	original := delArgs{ProductID: 1, Name: "x"}

	decoded, err := codec.Decode(mustEncode(t, codec, original))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded != original {
		t.Fatalf("round trip = %+v, want %+v", decoded, original)
	}
}

func mustEncode[A any](t *testing.T, codec Codec[A], args A) []byte {
	t.Helper()

	encoded, err := codec.Encode(args)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	return encoded
}

func TestRouteEncodeEnforcesSizeLimit(t *testing.T) {
	route := Route[string]{menu: "menu", name: "r", codec: positionalCodec[string]{}}

	if _, err := route.encode(strings.Repeat("x", 100)); err == nil {
		t.Fatal("expected callback_data_too_long error, got nil")
	}
}

func TestRouteRejectsInvalidNames(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid route name")
		}
	}()

	InlineMenuFor(NewInlineMenuRef("menu"), S("hi")).
		Route("bad:name", func(ctx *Ctx, args NoData) Action { return Stop() })
}

func TestInlineRouteDispatch(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	menuRef := NewInlineMenuRef("products")

	var got delArgs

	builder := InlineMenuFor(menuRef, S("hi"))
	route := builder.Route("del", func(ctx *Ctx, args delArgs) Action {
		got = args
		return Stop()
	})

	engine.Add(builder)

	// Mint a button and extract its callback data from the rendered markup.
	button := Do(S("Delete"), route, delArgs{ProductID: 42, Name: "a:b", InStock: true})
	if button.encodeErr != nil {
		t.Fatalf("encode button: %v", button.encodeErr)
	}

	ctx := NewTestUpdate().WithCallbackQuery(1, button.callbackData).BuildCtx("Welcome")
	ctx.engine = engine

	engine.processCallbackQuery(ctx)

	if got != (delArgs{ProductID: 42, Name: "a:b", InStock: true}) {
		t.Fatalf("handler received %+v, want the encoded payload", got)
	}
}

func TestGlobalRouteDispatch(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	var got int64

	route := engine.Route("track", func(ctx *Ctx, id int64) Action {
		got = id
		return Stop()
	})

	button := Do(S("Track"), route, 99)

	ctx := NewTestUpdate().WithCallbackQuery(1, button.callbackData).BuildCtx("Welcome")
	ctx.engine = engine

	engine.processCallbackQuery(ctx)

	if got != 99 {
		t.Fatalf("handler received %d, want 99", got)
	}
}

func TestUnknownCallbackReportsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	engine := New(NewMockUserRepository(), stateWelcome, NewOptions().SetLogger(logger))

	ctx := NewTestUpdate().WithCallbackQuery(1, "unknown:thing").BuildCtx("Welcome")
	ctx.engine = engine

	engine.processCallbackQuery(ctx)

	if len(hook.Entries) == 0 || !strings.Contains(hook.LastEntry().Message, "callback_query_handler_not_found") {
		t.Fatalf("expected callback_query_handler_not_found error, got %v", hook.Entries)
	}
}

func TestNoDataRouteDispatch(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	called := false

	builder := InlineMenuFor(NewInlineMenuRef("menu"), S("hi"))
	route := builder.Route("refresh", func(ctx *Ctx, _ NoData) Action {
		called = true
		return Stop()
	})

	engine.Add(builder)

	button := Do(S("Refresh"), route, NoData{})
	if button.encodeErr != nil {
		t.Fatalf("encode NoData payload: %v", button.encodeErr)
	}

	ctx := NewTestUpdate().WithCallbackQuery(1, button.callbackData).BuildCtx("Welcome")
	ctx.engine = engine

	engine.processCallbackQuery(ctx)

	if !called {
		t.Fatal("NoData route handler was not called")
	}
}

func TestMalformedInternalCallbackReportsError(t *testing.T) {
	logger, hook := logrustest.NewNullLogger()

	engine := New(NewMockUserRepository(), stateWelcome, NewOptions().SetLogger(logger))
	engine.Add(InlineMenuFor(NewInlineMenuRef("menu"), S("hi")))

	ctx := NewTestUpdate().WithCallbackQuery(1, "menu:@m").BuildCtx("Welcome")
	ctx.engine = engine

	engine.processCallbackQuery(ctx)

	if len(hook.Entries) == 0 || !strings.Contains(hook.LastEntry().Message, "malformed_callback") {
		t.Fatalf("expected malformed_callback error, got %v", hook.Entries)
	}
}

// === Inline keyboard rendering ===

func TestInlineConditionalRendering(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	runtime := &inlineMenuRuntime{
		name: "menu",
		buttons: []*InlineButton{
			Alert(S("Secret"), "hidden").If(func(*Ctx) bool { return false }),
			Alert(S("Shown"), "hi"),
		},
	}

	markup, err := engine.buildInlineMarkup(testCtx("hi"), runtime)
	if err != nil {
		t.Fatalf("buildInlineMarkup: %v", err)
	}

	rows := inlineRows(t, markup)
	if len(rows) != 1 || len(rows[0]) != 1 || rows[0][0].Text != "Shown" {
		t.Fatalf("hidden inline button rendered: %+v", rows)
	}

	if rows[0][0].CallbackData != "menu:@a:0:hi" {
		t.Fatalf("alert callback_data = %q, want %q", rows[0][0].CallbackData, "menu:@a:0:hi")
	}

	// All buttons hidden -> nil markup instead of an empty keyboard.
	hidden := &inlineMenuRuntime{
		name:    "menu",
		buttons: []*InlineButton{Alert(S("Secret"), "x").If(func(*Ctx) bool { return false })},
	}
	if markup, _ := engine.buildInlineMarkup(testCtx("hi"), hidden); markup != nil {
		t.Fatal("expected nil markup when all buttons are hidden")
	}
}

func TestInlineMarkupEnforcesSizeLimit(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	runtime := &inlineMenuRuntime{
		name:    "menu",
		buttons: []*InlineButton{Alert(S("A"), strings.Repeat("x", 100))},
	}

	if _, err := engine.buildInlineMarkup(testCtx("hi"), runtime); err == nil {
		t.Fatal("expected callback_data_too_long error, got nil")
	}
}

func TestInlineMarkupSkipsBrokenButtons(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	runtime := &inlineMenuRuntime{
		name: "menu",
		buttons: []*InlineButton{
			Alert(S("Broken"), strings.Repeat("x", 100)),
			Alert(S("Good"), "hi"),
		},
	}

	markup, err := engine.buildInlineMarkup(testCtx("hi"), runtime)
	if err == nil {
		t.Fatal("expected the broken button to be reported, got nil error")
	}
	if markup == nil {
		t.Fatal("expected markup with the good button despite the broken one")
	}

	rows := inlineRows(t, markup)
	if len(rows) != 1 || len(rows[0]) != 1 || rows[0][0].Text != "Good" {
		t.Fatalf("expected only the good button to render, got %+v", rows)
	}
}

// === State transitions and payloads ===

func TestGoToWithDeliversPayload(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	engine.Add(
		Menu(stateWelcome, nil),
		Menu(stateCheckout, nil),
	)

	ctx := testCtx("hi")
	ctx.engine = engine

	payload := checkoutData{ProductID: 7, Qty: 2}

	if err := engine.switchState(ctx, stateCheckout.Name(), &payload); err != nil {
		t.Fatalf("switchState: %v", err)
	}

	if ctx.State != "Checkout" || !ctx.IsSwitched {
		t.Fatalf("ctx.State = %q IsSwitched = %v", ctx.State, ctx.IsSwitched)
	}

	got, ok := StateData(ctx, stateCheckout)
	if !ok || got != payload {
		t.Fatalf("StateData = %+v, %v; want %+v, true", got, ok, payload)
	}

	// Wrong handle / wrong state must report false.
	if _, ok := StateData(ctx, stateWelcome); ok {
		t.Fatal("StateData with mismatched state should report false")
	}
}

func TestButtonStateTransition(t *testing.T) {
	repo := NewMockUserRepository()
	engine := New(repo, stateWelcome)

	engine.Add(
		Menu(stateWelcome, nil).Buttons(GoToWith(S("Buy"), stateCheckout, checkoutData{ProductID: 5})),
		Menu(stateCheckout, nil),
	)

	ctx := testCtx("Buy")
	ctx.engine = engine

	engine.processStateMenu(ctx, engine.getMenu("Welcome"))

	if ctx.State != "Checkout" {
		t.Fatalf("ctx.State = %q, want Checkout", ctx.State)
	}

	got, ok := StateData(ctx, stateCheckout)
	if !ok || got.ProductID != 5 {
		t.Fatalf("StateData = %+v, %v; want ProductID 5", got, ok)
	}

	if !repo.AssertStateSet(1, "Checkout") {
		t.Fatal("state transition was not persisted")
	}
}

func TestButtonHook(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	engine.Add(Menu(stateWelcome, nil), Menu(stateCheckout, nil))

	ctx := testCtx("hi")
	ctx.engine = engine

	// Hook returning Next lets the action proceed.
	proceed := GoTo(S("Go"), stateWelcome)
	proceed.Hook(func(ctx *Ctx) Action { return Next() })

	if !engine.runButtonHook(ctx, proceed) {
		t.Fatal("hook returning Next should let the action proceed")
	}

	// Hook returning Stop blocks the action.
	blocked := GoTo(S("Go"), stateWelcome)
	blocked.Hook(func(ctx *Ctx) Action { return Stop() })

	if engine.runButtonHook(ctx, blocked) {
		t.Fatal("hook returning Stop should block the action")
	}

	// Hook returning a transition runs it instead of the action.
	redirecting := GoTo(S("Go"), stateWelcome)
	redirecting.Hook(func(ctx *Ctx) Action { return ctx.GoTo(stateWelcome) })

	if engine.runButtonHook(ctx, redirecting) {
		t.Fatal("hook returning an action should block the button action")
	}
}

// === Part handlers (On[P]) ===

func TestOnPartDispatch(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	var gotPhoto []structs.PhotoSize

	runtime := Menu(stateWelcome, nil).
		On(func(ctx *Ctx, _ *NoData, photo []structs.PhotoSize) Action {
			gotPhoto = photo
			return Stop()
		}).
		compile()

	ctx := NewTestUpdate().WithPhoto("file_id_1").BuildCtx("Welcome")
	ctx.engine = engine

	engine.processStateMenu(ctx, runtime)

	if len(gotPhoto) != 1 || gotPhoto[0].FileId != "file_id_1" {
		t.Fatalf("photo handler received %+v", gotPhoto)
	}
}

func TestDefaultHandlerCatchesUnmatchedParts(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)

	defaultCalled := false

	runtime := Menu(stateWelcome, nil).
		On(func(ctx *Ctx, _ *NoData, photo []structs.PhotoSize) Action { return Stop() }).
		Default(func(ctx *Ctx, _ *NoData) Action {
			defaultCalled = true
			return Stop()
		}).
		compile()

	ctx := NewTestUpdate().WithDocument("doc_id", "doc.txt").BuildCtx("Welcome")
	ctx.engine = engine

	engine.processStateMenu(ctx, runtime)

	if !defaultCalled {
		t.Fatal("document message should fall back to the default handler")
	}
}

// === Multi-language label dispatch ===

func TestCrossLanguageLabelDispatch(t *testing.T) {
	languages, err := NewLanguageBuilder(language.English).
		RegisterTomlFormat([]string{"testdata/locale.en.toml", "testdata/locale.fa.toml"}).
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	engine := New(NewMockUserRepository(), stateWelcome).
		WithLanguageConfig(NewLanguageConfig(languages, NewMockUserLanguageRepository()))

	runtime := Menu(stateWelcome, S("hi")).Buttons(Reply(L("Greeting"), S("ok"))).compile()

	ctx := testCtx("hi")
	ctx.engine = engine
	ctx.SetLanguage(languages.GetByTag("fa"))

	_, byLabel, err := engine.renderReplyKeyboard(ctx, runtime)
	if err != nil {
		t.Fatalf("renderReplyKeyboard: %v", err)
	}

	if _, ok := byLabel["Salam"]; !ok {
		t.Fatal("current-language label missing from dispatch map")
	}

	// The user may press a keyboard sent before a language switch: the old
	// language's label must still dispatch.
	if _, ok := byLabel["Hello"]; !ok {
		t.Fatal("previous-language label missing from dispatch map")
	}
}

// === Engine internals ===

type transientErrorRepo struct {
	setStateCalled bool
}

func (r *transientErrorRepo) UpsertUser(user *structs.User) error { return nil }
func (r *transientErrorRepo) SetUserState(id int64, state string) error {
	r.setStateCalled = true
	return nil
}
func (r *transientErrorRepo) GetUserState(id int64) (string, error) {
	return "", errors.New("database is down")
}

func TestProcessUserStateTransientErrorDoesNotReset(t *testing.T) {
	repo := &transientErrorRepo{}
	engine := New(repo, stateWelcome)

	update := NewTestUpdate().WithMessage(1, 1, "hi").Build()

	if _, err := engine.processUserState(update); err == nil {
		t.Fatal("expected transient repository error to surface, got nil")
	}
	if repo.setStateCalled {
		t.Fatal("user state was reset on a transient repository error")
	}
}

func TestProcessUserStateNewUserGetsDefaultState(t *testing.T) {
	repo := NewMockUserRepository()
	engine := New(repo, stateWelcome)

	update := NewTestUpdate().WithMessage(1, 1, "hi").Build()

	state, err := engine.processUserState(update)
	if err != nil {
		t.Fatalf("new user: unexpected error: %v", err)
	}
	if state != "Welcome" {
		t.Fatalf("new user state = %q, want %q", state, "Welcome")
	}
	if !repo.AssertStateSet(1, "Welcome") {
		t.Fatal("default state was not persisted for new user")
	}
}

func TestValidate(t *testing.T) {
	stateMissing := NewState[NoData]("Missing")
	menuRef := NewInlineMenuRef("info")

	// Default state without a menu.
	engine := New(NewMockUserRepository(), stateWelcome)
	if err := engine.Validate(); err == nil {
		t.Fatal("expected no_menu_for_default_state error, got nil")
	}

	// Static button referencing an unregistered state.
	engine = New(NewMockUserRepository(), stateWelcome)
	engine.Add(Menu(stateWelcome, nil).Buttons(GoTo(S("Ghost"), stateMissing)))
	if err := engine.Validate(); err == nil {
		t.Fatal("expected no_menu_for_state error, got nil")
	}

	// Valid cross-references.
	engine = New(NewMockUserRepository(), stateWelcome)
	engine.Add(
		Menu(stateWelcome, nil).Buttons(ShowInline(S("Info"), menuRef), GoTo(S("Checkout"), stateWelcome)),
		InlineMenuFor(menuRef, S("info")).Buttons(OpenMenu(S("self"), menuRef)),
	)
	if err := engine.Validate(); err != nil {
		t.Fatalf("valid engine rejected: %v", err)
	}
}

func TestDuplicateRegistrationPanics(t *testing.T) {
	engine := New(NewMockUserRepository(), stateWelcome)
	engine.Add(Menu(stateWelcome, nil))

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for duplicate state menu")
		}
	}()

	engine.Add(Menu(stateWelcome, nil))
}

// === Languages and texts (unchanged behavior) ===

func TestLanguageFallsBackToDefaultTag(t *testing.T) {
	langs, err := NewLanguageBuilder(language.English).
		RegisterTomlFormat([]string{"testdata/locale.en.toml", "testdata/locale.fa.toml"}).
		Build()
	if err != nil {
		t.Fatalf("build languages: %v", err)
	}

	fa := langs.GetByTag("fa")
	if fa == nil {
		t.Fatal("fa language not found")
	}

	if got, err := fa.Get("Greeting"); err != nil || got != "Salam" {
		t.Fatalf("fa Greeting = %q, %v; want %q, nil", got, err, "Salam")
	}

	// Key missing in fa must fall back to the default language. go-i18n
	// returns the fallback text together with a "not found in fa" error,
	// so the text is what matters.
	if got, _ := fa.Get("OnlyEnglish"); got != "Only English" {
		t.Fatalf("fa OnlyEnglish = %q, want %q (fallback)", got, "Only English")
	}
}

func TestLocalizedTextFallbacks(t *testing.T) {
	ctx := testCtx("hi") // no language set

	if got := L("Nav.Home")(ctx); got != "Nav.Home" {
		t.Fatalf("L without language = %q, want the key %q", got, "Nav.Home")
	}

	if got := LP("Nav.Params", map[string]interface{}{"X": 1})(ctx); got != "Nav.Params" {
		t.Fatalf("LP without language = %q, want the key %q", got, "Nav.Params")
	}
}

func TestParseCommandEdgeCases(t *testing.T) {
	if cmd := ParseCommand("/"); cmd != nil {
		t.Fatalf("ParseCommand(\"/\") = %+v, want nil", cmd)
	}
	if cmd := ParseCommand("/@somebot"); cmd != nil {
		t.Fatalf("ParseCommand(\"/@somebot\") = %+v, want nil", cmd)
	}
	if cmd := ParseCommand("plain text"); cmd != nil {
		t.Fatalf("ParseCommand(non-command) = %+v, want nil", cmd)
	}

	cmd := ParseCommand("/start a b")
	if cmd == nil || cmd.Name != "start" || cmd.ArgCount() != 2 || cmd.Arg(1) != "b" || cmd.RawArgs != "a b" {
		t.Fatalf("ParseCommand(\"/start a b\") = %+v", cmd)
	}

	cmd = ParseCommand("/help@mybot x")
	if cmd == nil || cmd.Name != "help" || cmd.BotName != "mybot" {
		t.Fatalf("ParseCommand(\"/help@mybot x\") = %+v", cmd)
	}
}

func TestMockRepositoriesMatchRealContracts(t *testing.T) {
	if _, err := NewMockUserRepository().GetUserState(1); !errors.Is(err, UserStateNotFoundErr) {
		t.Fatalf("mock user repo GetUserState = %v, want UserStateNotFoundErr", err)
	}
	if _, err := NewMockUserLanguageRepository().GetUserLanguage(1); !errors.Is(err, UserLanguageNotFoundErr) {
		t.Fatalf("mock language repo GetUserLanguage = %v, want UserLanguageNotFoundErr", err)
	}
	if _, err := NewDefaultUserRepository().GetUserState(1); !errors.Is(err, UserStateNotFoundErr) {
		t.Fatalf("default user repo GetUserState = %v, want UserStateNotFoundErr", err)
	}
}
