package telejoon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api"
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
		// go-i18n returns fallback-language text together with an error, so
		// only non-empty text counts as a translation.
		translated, _ := lang.Get(fmt.Sprintf("%s.Text", cfg.changeLanguageState))
		if translated == "" {
			continue
		}

		text += fmt.Sprintf("%s\n", translated)
	}

	if text == "" {
		text = cfg.changeLanguageState + "\n"
	}

	languageLabel := func(lang Language) string {
		label, _ := lang.Get(fmt.Sprintf("%s.Button", cfg.changeLanguageState))
		if label == "" {
			label = lang.tag
		}

		return label
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

				if err := cfg.repo.SetUserLanguage(ctx.UserID(), lang.tag); err != nil {
					return Error(err)
				}

				chosen := lang
				ctx.SetLanguage(&chosen)

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
func (e *Engine) Process(client *tgbotapi.TelegramBot, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			if panicHandler := e.getPanicHandler(); panicHandler != nil {
				panicHandler(client, update, r, string(debug.Stack()))
			} else {
				e.onErr(client, update, fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
			}
		}
	}()

	userState, err := e.processUserState(update)
	if err != nil {
		e.onErr(client, update, err)
		return
	}

	ctx := &Ctx{
		client:  client,
		engine:  e,
		session: &sync.Map{},
		State:   userState,
		Update:  update,
	}

	from := update.From()
	if from == nil {
		j, _ := json.Marshal(update)
		e.onErr(client, update, fmt.Errorf("update.From() is nil: %s", string(j)))

		return
	}

	ctx.userID = from.Id

	var lang *Language

	languageConfig := e.getLanguageConfig()

	if languageConfig != nil {
		userLanguage, err := languageConfig.repo.GetUserLanguage(from.Id)

		switch {
		case err == nil:
			lang = languageConfig.languages.GetByTag(userLanguage)
		case errors.Is(err, UserLanguageNotFoundErr):
			// New user without a chosen language.
			if languageConfig.forceChooseLanguage && languageConfig.changeLanguageState != "" {
				if update.CallbackQuery != nil {
					go func() {
						_, err := client.Send(client.AnswerCallbackQuery().
							SetCallbackQueryId(update.CallbackQuery.Id).
							SetShowAlert(false))
						if err != nil {
							e.onErr(client, update, err)
						}
					}()
				}

				if userState != languageConfig.changeLanguageState {
					if err := e.switchState(ctx, languageConfig.changeLanguageState, nil); err != nil {
						e.onErr(client, update, err)
					}

					return
				}
			}
			// Otherwise proceed without a language; texts fall back gracefully.
		default:
			// Real repository error (e.g. database down): surface it instead of
			// silently misbehaving.
			e.onErr(client, update, fmt.Errorf("get_user_language: %w", err))

			return
		}
	}

	ctx.language = lang

	if !e.runMiddlewares(ctx, e.getMiddlewares()) {
		return
	}

	if update.Message != nil {
		if runtime := e.getMenu(userState); runtime != nil {
			e.processStateMenu(ctx, runtime)

			return
		}

		// The persisted state has no registered menu (e.g. left over from a
		// previous deploy): report it and reset the user to the default state.
		e.onErr(client, update, fmt.Errorf("no_handler_for_state: %s", userState))

		if userState != e.defaultState {
			if err := e.switchState(ctx, e.defaultState, nil); err != nil {
				e.onErr(client, update, err)
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
func (e *Engine) SwitchUserState(client *tgbotapi.TelegramBot, userID int64, state State[NoData]) error {
	lang, err := e.userLanguage(userID)
	if err != nil {
		return err
	}

	ctx := &Ctx{
		client:     client,
		engine:     e,
		session:    &sync.Map{},
		State:      state.name,
		language:   lang,
		IsSwitched: true,
		userID:     userID,
	}

	return e.switchState(ctx, state.name, nil)
}

// SendInlineMenu sends (or edits into) an inline menu for the given update.
func (e *Engine) SendInlineMenu(
	client *tgbotapi.TelegramBot,
	update tgbotapi.Update,
	menu InlineMenuRef,
	edit bool,
) error {

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
		e.onErr(ctx.client, ctx.Update, err)
	}
}

// processStateMenu runs the menu middlewares, dispatches the message to a
// button / text / part handler, and renders the menu when nothing stopped it.
func (e *Engine) processStateMenu(ctx *Ctx, runtime *menuRuntime) {
	if !e.runMiddlewares(ctx, runtime.middlewares) {
		return
	}

	ctx.stateData = runtime.loadData(ctx)

	markup, byLabel, err := e.renderReplyKeyboard(ctx, runtime)
	if err != nil {
		e.onErr(ctx.client, ctx.Update, err)

		return
	}

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

	if replyText := runtime.text(ctx); replyText != "" {
		_, err := ctx.client.Send(ctx.client.Message().
			SetText(replyText).
			SetChatId(ctx.UserID()).
			SetReplyMarkup(markup))
		if err != nil {
			e.onErr(ctx.client, ctx.Update,
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
			e.onErr(ctx.client, ctx.Update, fmt.Errorf("empty_callback_action: %s", runtime.name))

			return
		}

		if !e.runMiddlewares(ctx, runtime.middlewares) {
			return
		}

		switch parts[1] {
		case "@a": // alert button
			if len(parts) < 4 {
				e.onErr(ctx.client, ctx.Update, fmt.Errorf("malformed_alert_callback: %s", data))

				return
			}

			text, err := url.QueryUnescape(parts[3])
			if err != nil {
				e.onErr(ctx.client, ctx.Update, fmt.Errorf("malformed_alert_callback: %s", data))

				return
			}

			_, err = ctx.client.Send(ctx.client.AnswerCallbackQuery().
				SetCallbackQueryId(ctx.Update.CallbackQuery.Id).
				SetText(text).
				SetShowAlert(parts[2] == "1"))
			if err != nil {
				e.onErr(ctx.client, ctx.Update, err)
			}
		case "@m", "@e", "@s": // open menu / edit into menu / switch state
			if len(parts) < 3 || !validRouteName(parts[2]) {
				e.onErr(ctx.client, ctx.Update, fmt.Errorf("malformed_callback: %s", data))

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
				e.onErr(ctx.client, ctx.Update,
					fmt.Errorf("handler_for_action_not_found: %s", data))

				return
			}

			payload := ""
			if len(parts) > 2 {
				payload = parts[2]
			}

			e.processAction(ctx, asResult(handler(ctx, payload)))
		}

		return
	}

	if handler := e.getGlobalRoute(parts[0]); handler != nil {
		payload := ""
		if len(parts) > 1 {
			payload = parts[1]
		}

		e.processAction(ctx, asResult(handler(ctx, payload)))

		return
	}

	e.onErr(ctx.client, ctx.Update, fmt.Errorf("callback_query_handler_not_found: %s", parts[0]))
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

	if !e.runMiddlewares(ctx, runtime.middlewares) {
		return nil
	}

	markup, err := e.buildInlineMarkup(ctx, runtime)
	if err != nil {
		// Report, but still render: the markup simply omits the buttons that
		// failed to encode, instead of failing the whole menu silently.
		e.onErr(ctx.client, ctx.Update, err)
	}

	if runtime.text == nil {
		return fmt.Errorf("inline_menu_reply_text_not_set: %s", name)
	}

	replyText := runtime.text(ctx)
	if replyText == "" {
		return fmt.Errorf("inline_menu_reply_text_not_set: %s", name)
	}

	var cfg tgbotapi.Config

	if edit {
		if ctx.Update.CallbackQuery == nil || ctx.Update.CallbackQuery.Message == nil {
			return fmt.Errorf("cannot edit message: callback query or message is nil")
		}

		cfg = ctx.client.EditMessageText().SetText(replyText).
			SetChatId(ctx.UserID()).
			SetMessageId(ctx.Update.CallbackQuery.Message.MessageId).
			SetReplyMarkup(markup)
	} else {
		cfg = ctx.client.Message().
			SetText(replyText).
			SetChatId(ctx.UserID()).
			SetReplyMarkup(markup)
	}

	if _, err := ctx.client.Send(cfg); err != nil {
		return fmt.Errorf("error_sending_message_to_user: %d, %w", ctx.UserID(), err)
	}

	return nil
}

// answerCallback fires a no-op answer to the current callback query, stopping
// the client's loading spinner for internal navigation actions.
func (e *Engine) answerCallback(ctx *Ctx) {
	if ctx.Update.CallbackQuery == nil {
		return
	}

	go func() {
		_, err := ctx.client.Send(ctx.client.AnswerCallbackQuery().
			SetCallbackQueryId(ctx.Update.CallbackQuery.Id).
			SetShowAlert(false))
		if err != nil {
			e.onErr(ctx.client, ctx.Update, err)
		}
	}()
}

// maxStateSwitchDepth bounds chained state transitions within one request, so
// mutually redirecting menus cannot recurse unboundedly.
const maxStateSwitchDepth = 8

// switchState persists the new state (and its payload, when a
// StateDataRepository is configured) and re-enters processing for the target
// state's menu.
func (e *Engine) switchState(ctx *Ctx, state string, data any) error {
	if ctx.switchDepth >= maxStateSwitchDepth {
		return fmt.Errorf("state_switch_depth_exceeded: %s", state)
	}

	runtime := e.getMenu(state)
	if runtime == nil {
		return fmt.Errorf("no_handler_for_state: %s", state)
	}

	userID := ctx.UserID()

	if err := e.userRepository.SetUserState(userID, state); err != nil {
		return fmt.Errorf("error_setting_user_state: %d, %w", userID, err)
	}

	if data != nil {
		if repo := e.getStateDataRepository(); repo != nil {
			raw, err := json.Marshal(data)
			if err != nil {
				return fmt.Errorf("state_data_encode: %s: %w", state, err)
			}

			if err := repo.SetUserStateData(userID, state, raw); err != nil {
				return fmt.Errorf("state_data_store: %s: %w", state, err)
			}
		}
	}

	ctx.State = state
	ctx.IsSwitched = true
	ctx.pendingData = data
	ctx.stateData = nil
	// Memoized conditions were evaluated for the previous state's rendering;
	// they must not leak into the new state's.
	ctx.condResults = nil
	ctx.switchDepth++

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
	}

	return lang, nil
}
