package telejoon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// Engine is the private-chat state machine: it resolves the user's state,
// language, and typed state payload, then dispatches the update to the
// matching menu. It is the monomorphic hub of the framework — all generics
// live at the registration edge and erase into closures at Add.
type Engine struct {
	engine

	userRepository UserRepository

	stateDataRepository StateDataRepository

	m sync.RWMutex

	panicHandler PanicHandler

	middlewares []Handler

	defaultState string

	menus map[string]*menuRuntime

	inlineMenus map[string]*inlineMenuRuntime

	globalRoutes map[string]erasedRouteHandler

	languageConfig *LanguageConfig
}

// New creates an Engine with the given user repository and default (initial)
// state. The default state must be payload-less.
func New(userRepo UserRepository, defaultState State[NoData], opts ...*Options) *Engine {
	return &Engine{
		engine:         newEngine(opts...),
		userRepository: userRepo,
		defaultState:   defaultState.name,
		menus:          map[string]*menuRuntime{},
		inlineMenus:    map[string]*inlineMenuRuntime{},
		globalRoutes:   map[string]erasedRouteHandler{},
	}
}

// Add compiles menus (*MenuBuilder[D], *InlineMenuBuilder) into the engine.
// It panics on structural registration errors (duplicate or invalid names),
// which are startup misconfigurations.
func (e *Engine) Add(items ...Registrable) *Engine {
	e.m.Lock()
	defer e.m.Unlock()

	for _, item := range items {
		if err := item.register(e); err != nil {
			panic(fmt.Sprintf("telejoon: %s", err))
		}
	}

	return e
}

// Use adds a global middleware that runs before any menu processing.
// Return Next to continue, anything else to stop.
func (e *Engine) Use(middleware Handler) *Engine {
	e.m.Lock()
	defer e.m.Unlock()

	e.middlewares = append(e.middlewares, middleware)

	return e
}

// WithPanicHandler sets the panic handler called on recovered panics.
func (e *Engine) WithPanicHandler(handler PanicHandler) *Engine {
	e.m.Lock()
	defer e.m.Unlock()

	e.panicHandler = handler

	return e
}

// WithStateDataRepository enables persistence of state payloads.
func (e *Engine) WithStateDataRepository(repo StateDataRepository) *Engine {
	e.m.Lock()
	defer e.m.Unlock()

	e.stateDataRepository = repo

	return e
}

// Route registers a global typed callback handler (not bound to any inline
// menu) and returns the typed handle used to mint buttons with Do.
func (e *Engine) Route[A any](name string, fn func(ctx *Ctx, args A) Action, opts ...RouteOption[A]) Route[A] {
	if !validRouteName(name) {
		panic(fmt.Sprintf("telejoon: invalid route name %q", name))
	}

	e.m.Lock()
	defer e.m.Unlock()

	if _, exists := e.globalRoutes[name]; exists {
		panic(fmt.Sprintf("telejoon: duplicate global route %q", name))
	}

	if _, conflict := e.inlineMenus[name]; conflict {
		panic(fmt.Sprintf("telejoon: global route %q conflicts with an inline menu", name))
	}

	route := Route[A]{name: name, codec: positionalCodec[A]{}}
	for _, opt := range opts {
		opt(&route)
	}

	e.globalRoutes[name] = route.erase(fn)

	return route
}

// WithLanguageConfig adds a language config to the engine and registers the
// change-language menu when configured.
func (e *Engine) WithLanguageConfig(cfg *LanguageConfig) *Engine {
	e.m.Lock()
	defer e.m.Unlock()

	e.languageConfig = cfg

	if cfg.changeLanguageState == "" || cfg.languages == nil {
		return e
	}

	text := ""

	for _, lang := range cfg.languages.localizers {
		// Exact-language lookup: a language missing the chooser text must
		// not contribute a duplicate copy of the default language's line.
		translated, ok := lang.GetOwn(fmt.Sprintf("%s.Text", cfg.changeLanguageState))
		if !ok {
			continue
		}

		text += fmt.Sprintf("%s\n", translated)
	}

	if text == "" {
		text = cfg.changeLanguageState + "\n"
	}

	languageLabel := func(lang Language) string {
		// Exact-language lookup: a missing translation falls back to the
		// language tag, never to the default language's label (which would
		// also trip the duplicate-label check below).
		label, ok := lang.GetOwn(fmt.Sprintf("%s.Button", cfg.changeLanguageState))
		if !ok {
			label = lang.tag
		}

		return label
	}

	// Chooser labels double as dispatch keys, so two languages translating to
	// the same label would make the second one unselectable. That is a
	// startup misconfiguration: panic loudly instead of misrouting at runtime.
	seenLabels := map[string]string{}

	for _, lang := range cfg.languages.localizers {
		label := languageLabel(lang)
		if other, dup := seenLabels[label]; dup {
			panic(fmt.Sprintf(
				"telejoon: change-language menu %q: languages %q and %q share the button label %q",
				cfg.changeLanguageState, other, lang.tag, label))
		}

		seenLabels[label] = lang.tag
	}

	menu := Menu(NewState[NoData](cfg.changeLanguageState), S(text)).
		ButtonsFunc(func(ctx *Ctx, _ *NoData) []*Button {
			var buttons []*Button

			for _, lang := range cfg.languages.localizers {
				buttons = append(buttons, Raw(S(languageLabel(lang))))
			}

			return buttons
		}).
		OnText(func(ctx *Ctx, _ *NoData, message string) Action {
			for _, lang := range cfg.languages.localizers {
				if message != languageLabel(lang) {
					continue
				}

				if err := ctx.ChangeLanguage(lang.tag); err != nil {
					return Error(err)
				}

				return actionResult{kind: actionKindState, state: e.defaultState}
			}

			return Next()
		})

	if err := menu.register(e); err != nil {
		panic(fmt.Sprintf("telejoon: %s", err))
	}

	return e
}

// Validate checks cross-references between registered menus: the default
// state, statically-known state targets, and inline menu targets.
// Buttons produced by ButtonsFunc are per-request and cannot be checked.
func (e *Engine) Validate() error {
	e.m.RLock()
	defer e.m.RUnlock()

	if _, ok := e.menus[e.defaultState]; !ok {
		return fmt.Errorf("no_menu_for_default_state: %s", e.defaultState)
	}

	for _, runtime := range e.menus {
		for _, button := range runtime.staticButtons {
			if button == nil {
				continue
			}

			switch button.kind {
			case buttonKindState:
				if _, ok := e.menus[button.state]; !ok {
					return fmt.Errorf("no_menu_for_state: %s (referenced from %s)", button.state, runtime.state)
				}
			case buttonKindInlineMenu:
				if _, ok := e.inlineMenus[button.menu]; !ok {
					return fmt.Errorf("no_inline_menu: %s (referenced from %s)", button.menu, runtime.state)
				}
			}
		}
	}

	for _, runtime := range e.inlineMenus {
		for _, button := range runtime.buttons {
			if button == nil {
				continue
			}

			switch button.kind {
			case inlineKindMenu:
				if _, ok := e.inlineMenus[button.menu]; !ok {
					return fmt.Errorf("no_inline_menu: %s (referenced from %s)", button.menu, runtime.name)
				}
			case inlineKindState:
				if _, ok := e.menus[button.state]; !ok {
					return fmt.Errorf("no_menu_for_state: %s (referenced from %s)", button.state, runtime.name)
				}
			}
		}
	}

	if e.languageConfig != nil && e.languageConfig.languages != nil {
		// Declared message handles (NewMsg) must localize with the default
		// language; a typo'd key fails here instead of leaking into a chat.
		if missing := e.languageConfig.languages.validateMsgs(); len(missing) > 0 {
			return fmt.Errorf("messages_not_translated: %s", strings.Join(missing, ", "))
		}
	}

	return nil
}

func (e *Engine) getMenu(state string) *menuRuntime {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.menus[state]
}

func (e *Engine) getInlineMenu(name string) (*inlineMenuRuntime, bool) {
	e.m.RLock()
	defer e.m.RUnlock()

	runtime, ok := e.inlineMenus[name]

	return runtime, ok
}

func (e *Engine) getGlobalRoute(name string) erasedRouteHandler {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.globalRoutes[name]
}

func (e *Engine) getLanguageConfig() *LanguageConfig {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.languageConfig
}

func (e *Engine) getStateDataRepository() StateDataRepository {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.stateDataRepository
}

func (e *Engine) getPanicHandler() PanicHandler {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.panicHandler
}

func (e *Engine) getMiddlewares() []Handler {
	e.m.RLock()
	defer e.m.RUnlock()

	return append([]Handler(nil), e.middlewares...)
}

func (e *Engine) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil && chat.Type == "private" {
		return true
	}

	return false
}

// Process handles a single private-chat update. Each update runs in its own
// goroutine (see Start), so panics are recovered per update.
func (e *Engine) Process(reqCtx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			e.reportPanic(reqCtx, client, update, r)
		}
	}()

	userState, err := e.processUserState(update)
	if err != nil {
		e.onErr(reqCtx, client, update, err)
		return
	}

	ctx := &Ctx{
		client:  client,
		engine:  e,
		reqCtx:  reqCtx,
		session: &sync.Map{},
		State:   userState,
		Update:  update,
	}

	// processUserState already rejected a nil From.
	ctx.userID = update.From().Id

	var lang *Language

	forceChoose := false

	languageConfig := e.getLanguageConfig()

	if languageConfig != nil {
		userLanguage, err := languageConfig.repo.GetUserLanguage(ctx.userID)

		switch {
		case err == nil:
			lang = languageConfig.languages.GetByTag(userLanguage)
			if lang == nil {
				// The stored tag is not among the configured languages
				// (e.g. a language was removed): fall back to the default
				// instead of rendering raw message keys.
				lang = languageConfig.languages.Default()
			}
		case errors.Is(err, UserLanguageNotFoundErr):
			// New user without a chosen language: redirect to the chooser,
			// but only after the global middlewares have run (below) — an
			// auth or ban middleware must not be bypassed by first contact.
			forceChoose = languageConfig.forceChooseLanguage &&
				languageConfig.changeLanguageState != "" &&
				userState != languageConfig.changeLanguageState
		default:
			// Real repository error (e.g. database down): surface it instead of
			// silently misbehaving.
			e.onErr(reqCtx, client, update, fmt.Errorf("get_user_language: %w", err))

			return
		}
	}

	ctx.language = lang

	if !e.runMiddlewares(ctx, e.getMiddlewares()) {
		return
	}

	// A global middleware may have assigned a language itself
	// (ctx.ChangeLanguage / SetLanguage); only redirect users still without
	// one.
	if forceChoose && ctx.language == nil {
		if update.CallbackQuery != nil {
			e.answerCallback(ctx)
		}

		if err := e.switchState(ctx, languageConfig.changeLanguageState, nil); err != nil {
			e.onErr(reqCtx, client, update, err)
		}

		return
	}

	if update.Message != nil {
		if runtime := e.getMenu(userState); runtime != nil {
			e.processStateMenu(ctx, runtime)

			return
		}

		// The persisted state has no registered menu (e.g. left over from a
		// previous deploy): report it and reset the user to the default state.
		e.onErr(reqCtx, client, update, fmt.Errorf("no_handler_for_state: %s", userState))

		if userState != e.defaultState {
			if err := e.switchState(ctx, e.defaultState, nil); err != nil {
				e.onErr(reqCtx, client, update, err)
			}
		}

		return
	}

	if update.CallbackQuery != nil {
		e.processCallbackQuery(ctx)
	}
}

// SwitchUserState transitions a user to a payload-less state from outside of
// update processing (e.g. a background job).
func (e *Engine) SwitchUserState(
	reqCtx context.Context,
	client *tgbotapi.Bot,
	userID int64,
	state State[NoData],
) (err error) {

	// Unlike Process this runs on the caller's goroutine, so recover panics
	// (e.g. from a menu's text builder) instead of crashing the caller.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			e.reportPanic(reqCtx, client, tgbotapi.Update{}, r)
		}
	}()

	lang, err := e.userLanguage(userID)
	if err != nil {
		return err
	}

	// Resolve the user's CURRENT state so switchState sees the real
	// transition source (and deletes its leftover payload).
	currentState, err := e.userRepository.GetUserState(userID)
	if err != nil && !errors.Is(err, UserStateNotFoundErr) {
		return fmt.Errorf("get_user_state: %w", err)
	}

	ctx := &Ctx{
		client:     client,
		engine:     e,
		reqCtx:     reqCtx,
		session:    &sync.Map{},
		State:      currentState,
		language:   lang,
		IsSwitched: true,
		userID:     userID,
	}

	return e.switchState(ctx, state.name, nil)
}

// SendInlineMenu sends (or edits into) an inline menu for the given update.
func (e *Engine) SendInlineMenu(
	reqCtx context.Context,
	client *tgbotapi.Bot,
	update tgbotapi.Update,
	menu InlineMenuRef,
	edit bool,
) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			e.reportPanic(reqCtx, client, update, r)
		}
	}()

	userState, err := e.processUserState(update)
	if err != nil {
		return err
	}

	var lang *Language

	if from := update.From(); from != nil {
		lang, err = e.userLanguage(from.Id)
		if err != nil {
			return err
		}
	}

	ctx := &Ctx{
		client:   client,
		engine:   e,
		reqCtx:   reqCtx,
		session:  &sync.Map{},
		State:    userState,
		language: lang,
		Update:   update,
		userID:   updateUserID(update),
	}

	return e.processInlineMenu(ctx, menu.name, edit)
}

func updateUserID(update tgbotapi.Update) int64 {
	if from := update.From(); from != nil {
		return from.Id
	}

	return 0
}

// asResult normalizes a handler result; a nil Action is treated as Stop.
func asResult(action Action) actionResult {
	if action == nil {
		return actionResult{kind: actionKindStop}
	}

	return action.(actionResult)
}

// runMiddlewares runs a middleware chain. It returns false when processing
// should stop (the middleware's non-Next result has been processed already).
func (e *Engine) runMiddlewares(ctx *Ctx, middlewares []Handler) bool {
	for _, middleware := range middlewares {
		if middleware == nil {
			continue
		}

		result := asResult(middleware(ctx))
		if result.kind == actionKindNext {
			continue
		}

		if result.kind != actionKindStop {
			e.processAction(ctx, result)
		}

		return false
	}

	return true
}

// runMenuMiddlewares runs a menu's middleware chain at most once per update.
// The menu is marked as entered BEFORE the chain runs: a chain that stops
// processing (Stop, or a transition like a middleware redirect) must not
// re-run when the update re-enters the menu — "at most once per menu per
// update" holds whether the chain completed or not.
func (e *Engine) runMenuMiddlewares(ctx *Ctx, key string, middlewares []Handler) bool {
	if len(middlewares) == 0 {
		return true
	}

	if ctx.enteredMenus == nil {
		ctx.enteredMenus = map[string]struct{}{}
	}

	if _, ok := ctx.enteredMenus[key]; ok {
		return true
	}

	ctx.enteredMenus[key] = struct{}{}

	return e.runMiddlewares(ctx, middlewares)
}

// processAction executes a non-control-flow action result.
func (e *Engine) processAction(ctx *Ctx, result actionResult) {
	var err error

	switch result.kind {
	case actionKindError:
		err = result.err
	case actionKindState:
		err = e.switchState(ctx, result.state, result.data)
	case actionKindInlineMenu:
		err = e.processInlineMenu(ctx, result.menu, result.edit)
	}

	if err != nil {
		e.onErr(ctx.Context(), ctx.client, ctx.Update, err)
	}
}

// processStateMenu runs the menu middlewares, dispatches the message to a
// button / text / part handler, and renders the menu when nothing stopped it.
func (e *Engine) processStateMenu(ctx *Ctx, runtime *menuRuntime) {
	if !e.runMenuMiddlewares(ctx, "s:"+runtime.state, runtime.middlewares) {
		return
	}

	stateData, err := runtime.loadData(ctx)
	if err != nil {
		// A payload that cannot be loaded or decoded must not reach handlers
		// as a silent zero value — a checkout handler placing an order for
		// product 0 is worse than a dropped update.
		e.onErr(ctx.Context(), ctx.client, ctx.Update, err)

		return
	}

	ctx.stateData = stateData

	markup, byLabel, err := e.renderReplyKeyboard(ctx, runtime)
	if err != nil {
		e.onErr(ctx.Context(), ctx.client, ctx.Update, err)

		return
	}

	// handlerRan tracks whether any handler consumed (and possibly mutated
	// things feeding the keyboard) during dispatch. When dispatch then falls
	// through to the render step, the pre-dispatch keyboard is stale and must
	// be re-rendered.
	handlerRan := false

	if !ctx.IsSwitched {
		if message := ctx.Update.Message; message != nil {
			if message.Text != "" {
				if button := byLabel[message.Text]; button != nil {
					if !e.runButtonHook(ctx, button) {
						return
					}

					switch button.kind {
					case buttonKindText:
						e.processAction(ctx, asResult(ctx.ReplyText(button.text(ctx))))

						return
					case buttonKindState:
						e.processAction(ctx, actionResult{
							kind:  actionKindState,
							state: button.state,
							data:  button.data,
						})

						return
					case buttonKindInlineMenu:
						e.processAction(ctx, actionResult{
							kind: actionKindInlineMenu,
							menu: button.menu,
						})

						return
					case buttonKindRaw:
						// Raw buttons have no built-in action; the press falls
						// through to OnText like any other text message.
					}
				}

				if runtime.onText != nil {
					handlerRan = true

					result := asResult(runtime.onText(ctx))
					if result.kind != actionKindNext {
						e.processAction(ctx, result)

						return
					}
				}
			}

			var handler func(*Ctx) Action

			if kind, ok := detectPart(message); ok && runtime.parts[kind] != nil {
				handler = runtime.parts[kind]
			} else if message.Text == "" || runtime.onText == nil {
				handler = runtime.onDefault
			}

			if handler != nil {
				handlerRan = true

				result := asResult(handler(ctx))
				if result.kind != actionKindNext {
					e.processAction(ctx, result)

					return
				}
			}
		}
	}

	if runtime.text == nil {
		return
	}

	if handlerRan {
		// A handler ran and fell through with Next: re-render the keyboard so
		// mutations it made (cart contents, toggled flags feeding ButtonsFunc
		// or When) are reflected in what is sent. Memoized conditions were
		// evaluated for the pre-dispatch render; clear them so visibility is
		// re-evaluated against the mutated state.
		ctx.condResults = nil

		markup, _, err = e.renderReplyKeyboard(ctx, runtime)
		if err != nil {
			e.onErr(ctx.Context(), ctx.client, ctx.Update, err)

			return
		}
	}

	if replyText := runtime.text(ctx); replyText != "" {
		message := ctx.client.Message().
			ChatID(ctx.UserID()).
			Text(replyText)

		if markup != nil {
			message.ReplyMarkup(markup)
		} else {
			// No visible buttons: explicitly remove any keyboard a previous
			// state left behind, or its stale buttons would keep dispatching
			// into this state's OnText as unmatched text.
			message.RemoveReplyKeyboard()
		}

		_, err := message.Send(ctx.Context())
		if err != nil {
			e.onErr(ctx.Context(), ctx.client, ctx.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", ctx.UserID(), err))
		}
	}
}

// runButtonHook runs a button's hook. It returns false when the button's
// action should be skipped (the hook's result has been processed already).
func (e *Engine) runButtonHook(ctx *Ctx, button *Button) bool {
	if button.hook == nil {
		return true
	}

	result := asResult(button.hook(ctx))
	if result.kind == actionKindNext {
		return true
	}

	if result.kind != actionKindStop {
		e.processAction(ctx, result)
	}

	return false
}

// processCallbackQuery dispatches a callback query to an inline menu route,
// an internal inline button action, or a global route.
func (e *Engine) processCallbackQuery(ctx *Ctx) {
	data := ctx.CallbackData()
	if data == "" {
		return
	}

	parts := strings.Split(data, ":")
	if parts[0] == "" {
		return
	}

	if runtime, ok := e.getInlineMenu(parts[0]); ok {
		if len(parts) < 2 {
			e.onErr(ctx.Context(), ctx.client, ctx.Update, fmt.Errorf("empty_callback_action: %s", runtime.name))

			return
		}

		if !e.runMenuMiddlewares(ctx, "i:"+runtime.name, runtime.middlewares) {
			return
		}

		switch parts[1] {
		case "@a": // alert button
			if len(parts) < 4 {
				e.onErr(ctx.Context(), ctx.client, ctx.Update, fmt.Errorf("malformed_alert_callback: %s", data))

				return
			}

			text, err := url.QueryUnescape(parts[3])
			if err != nil {
				e.onErr(ctx.Context(), ctx.client, ctx.Update, fmt.Errorf("malformed_alert_callback: %s", data))

				return
			}

			_, err = ctx.client.AnswerCallbackQuery().
				CallbackQueryID(ctx.Update.CallbackQuery.Id).
				Text(text).
				ShowAlert(parts[2] == "1").
				Send(ctx.Context())
			if err != nil {
				e.onErr(ctx.Context(), ctx.client, ctx.Update, err)
			}
		case "@m", "@e", "@s": // open menu / edit into menu / switch state
			if len(parts) < 3 || !validRouteName(parts[2]) {
				e.onErr(ctx.Context(), ctx.client, ctx.Update, fmt.Errorf("malformed_callback: %s", data))

				return
			}

			// Internal actions answer the callback query so the client's
			// loading spinner stops.
			e.answerCallback(ctx)

			switch parts[1] {
			case "@m":
				e.processAction(ctx, actionResult{kind: actionKindInlineMenu, menu: parts[2]})
			case "@e":
				e.processAction(ctx, actionResult{kind: actionKindInlineMenu, menu: parts[2], edit: true})
			case "@s":
				e.processAction(ctx, actionResult{kind: actionKindState, state: parts[2]})
			}
		default:
			handler := runtime.routes[parts[1]]
			if handler == nil {
				e.onErr(ctx.Context(), ctx.client, ctx.Update,
					fmt.Errorf("handler_for_action_not_found: %s", data))

				return
			}

			payload := ""
			if len(parts) > 2 {
				payload = parts[2]
			}

			e.processAction(ctx, asResult(handler(ctx, payload)))
			e.answerCallback(ctx)
		}

		return
	}

	if handler := e.getGlobalRoute(parts[0]); handler != nil {
		payload := ""
		if len(parts) > 1 {
			payload = parts[1]
		}

		e.processAction(ctx, asResult(handler(ctx, payload)))
		e.answerCallback(ctx)

		return
	}

	e.onErr(ctx.Context(), ctx.client, ctx.Update, fmt.Errorf("callback_query_handler_not_found: %s", parts[0]))
}

// processInlineMenu renders and sends (or edits into) an inline menu.
func (e *Engine) processInlineMenu(ctx *Ctx, name string, edit bool) error {
	runtime, ok := e.getInlineMenu(name)
	if !ok {
		return fmt.Errorf("inline_menu_not_found: %s", name)
	}

	if ctx.UserID() == 0 {
		return fmt.Errorf("cannot process inline menu %s: no user in context", name)
	}

	if !e.runMenuMiddlewares(ctx, "i:"+runtime.name, runtime.middlewares) {
		return nil
	}

	markup, err := e.buildInlineMarkup(ctx, runtime)
	if err != nil {
		// Report, but still render: the markup simply omits the buttons that
		// failed to encode, instead of failing the whole menu silently.
		e.onErr(ctx.Context(), ctx.client, ctx.Update, err)
	}

	if runtime.text == nil {
		return fmt.Errorf("inline_menu_reply_text_not_set: %s", name)
	}

	replyText := runtime.text(ctx)
	if replyText == "" {
		return fmt.Errorf("inline_menu_reply_text_not_set: %s", name)
	}

	if edit {
		if ctx.Update.CallbackQuery == nil || ctx.Update.CallbackQuery.Message == nil {
			return fmt.Errorf("cannot edit message: callback query or message is nil")
		}

		_, err = ctx.client.EditMessageText().
			Text(replyText).
			ChatID(ctx.UserID()).
			MessageID(ctx.Update.CallbackQuery.Message.MessageId).
			ReplyMarkup(markup).
			Send(ctx.Context())
	} else {
		_, err = ctx.client.Message().
			ChatID(ctx.UserID()).
			Text(replyText).
			ReplyMarkup(markup).
			Send(ctx.Context())
	}

	if err != nil {
		// Re-rendering an unchanged menu (the refresh idiom with nothing new
		// to show) is rejected by Telegram with "message is not modified" —
		// that is a no-op, not an error.
		if edit && strings.Contains(err.Error(), "message is not modified") {
			return nil
		}

		return fmt.Errorf("error_sending_message_to_user: %d, %w", ctx.UserID(), err)
	}

	return nil
}

// answerCallback fires a no-op answer to the current callback query, stopping
// the client's loading spinner. It is a no-op when the query was already
// answered (via ctx.AnswerCallback or an earlier internal answer), so route
// handlers that forget to answer never leave the spinner running. It is
// synchronous: an untracked goroutine could be dropped mid-flight during
// shutdown.
func (e *Engine) answerCallback(ctx *Ctx) {
	if ctx.Update.CallbackQuery == nil || ctx.callbackAnswered || ctx.client == nil {
		return
	}

	ctx.callbackAnswered = true

	_, err := ctx.client.AnswerCallbackQuery().
		CallbackQueryID(ctx.Update.CallbackQuery.Id).
		ShowAlert(false).
		Send(ctx.Context())
	if err != nil {
		e.onErr(ctx.Context(), ctx.client, ctx.Update, err)
	}
}

// maxStateSwitchDepth bounds chained state transitions within one request, so
// mutually redirecting menus cannot recurse unboundedly.
const maxStateSwitchDepth = 8

// switchState persists the new state (and its payload, when a
// StateDataRepository is configured) and re-enters processing for the target
// state's menu.
//
// A payload lives exactly as long as the user continuously occupies its
// state. To enforce that, the write order is: encode and store the payload
// (or clear the target's leftover payload on a payload-less cross-state
// entry), publish the state name, then delete the previous state's payload.
// A stored payload is rolled back when publishing fails, so a failure at
// any step leaves neither a user stuck in a state without its payload nor
// an orphaned payload waiting to resurface.
func (e *Engine) switchState(ctx *Ctx, state string, data any) error {
	if ctx.switchDepth >= maxStateSwitchDepth {
		return fmt.Errorf("state_switch_depth_exceeded: %s", state)
	}

	runtime := e.getMenu(state)
	if runtime == nil {
		return fmt.Errorf("no_handler_for_state: %s", state)
	}

	userID := ctx.UserID()

	previousState := ctx.State
	crossState := previousState != "" && previousState != state

	repo := e.getStateDataRepository()

	var raw []byte

	if data != nil && repo != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("state_data_encode: %s: %w", state, err)
		}

		raw = encoded
	}

	if repo != nil {
		if raw != nil {
			if err := repo.SetUserStateData(userID, state, raw); err != nil {
				return fmt.Errorf("state_data_store: %s: %w", state, err)
			}
		} else if crossState {
			// A payload-less entry into a different state must never
			// resurrect data an earlier failed transition left behind:
			// clear the target first. Failure aborts the transition —
			// handlers must not see stale payloads. (Same-state re-entry
			// keeps the payload: that is the re-render pattern.)
			if err := repo.DeleteUserStateData(userID, state); err != nil {
				return fmt.Errorf("state_data_clear: %s: %w", state, err)
			}
		}
	}

	if err := e.userRepository.SetUserState(userID, state); err != nil {
		// Roll back a payload stored for a state the user never entered.
		if raw != nil {
			if derr := repo.DeleteUserStateData(userID, state); derr != nil {
				e.onErr(ctx.Context(), ctx.client, ctx.Update,
					fmt.Errorf("state_data_rollback: %s: %w", state, derr))
			}
		}

		return fmt.Errorf("error_setting_user_state: %d, %w", userID, err)
	}

	ctx.State = state
	ctx.IsSwitched = true
	ctx.pendingData = data
	ctx.stateData = nil
	// Memoized conditions were evaluated for the previous state's rendering;
	// they must not leak into the new state's.
	ctx.condResults = nil
	ctx.switchDepth++

	if repo != nil && crossState {
		if err := repo.DeleteUserStateData(userID, previousState); err != nil {
			// Payload hygiene, not transition correctness: report, don't fail.
			e.onErr(ctx.Context(), ctx.client, ctx.Update,
				fmt.Errorf("state_data_delete: %s: %w", previousState, err))
		}
	}

	e.processStateMenu(ctx, runtime)

	return nil
}

func (e *Engine) processUserState(update tgbotapi.Update) (string, error) {
	from := update.From()

	if from == nil {
		return "", errors.New("empty_from")
	}

	if err := e.userRepository.UpsertUser(from); err != nil {
		return "", fmt.Errorf("cant_store_user: %s", err)
	}

	if e.defaultState == "" {
		return "", fmt.Errorf("empty_default_state_name")
	}

	userState, err := e.userRepository.GetUserState(from.Id)
	if err != nil && !errors.Is(err, UserStateNotFoundErr) {
		// Real repository error (e.g. database down): surface it instead of
		// silently resetting the user to the default state.
		return "", fmt.Errorf("get_user_state: %w", err)
	}

	if userState == "" {
		userState = e.defaultState

		if err := e.userRepository.SetUserState(from.Id, userState); err != nil {
			return "", fmt.Errorf("store_user_state: %w", err)
		}
	}

	return userState, nil
}

func (e *Engine) userLanguage(userID int64) (*Language, error) {
	var lang *Language

	if languageConfig := e.getLanguageConfig(); languageConfig != nil {
		userLanguage, err := languageConfig.repo.GetUserLanguage(userID)
		if err != nil && !errors.Is(err, UserLanguageNotFoundErr) {
			return nil, fmt.Errorf("get_user_language: %w", err)
		}

		if userLanguage != "" {
			lang = languageConfig.languages.GetByTag(userLanguage)
		}

		if lang == nil {
			// No choice yet, or a stored tag that is no longer configured:
			// fall back to the default language.
			lang = languageConfig.languages.Default()
		}
	}

	return lang, nil
}

// reportPanic routes a recovered panic to the configured panic handler, or to
// the error handler when none is set.
func (e *Engine) reportPanic(reqCtx context.Context, client *tgbotapi.Bot, update tgbotapi.Update, r any) {
	if panicHandler := e.getPanicHandler(); panicHandler != nil {
		panicHandler(client, update, r, string(debug.Stack()))
	} else {
		e.onErr(reqCtx, client, update, fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
	}
}
