package telejoon

import (
	"errors"
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"strings"
	"sync"
)

type EngineWithPrivateStateHandlers[User any] struct {
	engine[User, any, any, any]

	userRepository UserRepository[User]

	m sync.Mutex

	middlewares []func(client *tgbotapi.TelegramBot, update *StateUpdate[User]) (string, bool)

	defaultStateName string

	staticMenus map[string]*StaticMenu[User]

	inlineMenus map[string]*InlineMenu[User]

	callbackQueryHandlers map[string]func(client *tgbotapi.TelegramBot, update *StateUpdate[User], args ...string) error

	languageConfig *LanguageConfig
}

func WithPrivateStateHandlers[User any](
	userRepo UserRepository[User], defaultState string, opts ...*Options) *EngineWithPrivateStateHandlers[User] {

	return &EngineWithPrivateStateHandlers[User]{
		engine: engine[User, any, any, any]{
			opts: opts,
		},
		userRepository:        userRepo,
		defaultStateName:      defaultState,
		staticMenus:           map[string]*StaticMenu[User]{},
		inlineMenus:           map[string]*InlineMenu[User]{},
		callbackQueryHandlers: map[string]func(*tgbotapi.TelegramBot, *StateUpdate[User], ...string) error{},
	}
}

// AddStaticMenu adds a static state handler
func (e *EngineWithPrivateStateHandlers[User]) AddStaticMenu(
	state string, handler *StaticMenu[User]) *EngineWithPrivateStateHandlers[User] {

	e.m.Lock()
	defer e.m.Unlock()

	e.staticMenus[state] = handler

	return e
}

func (e *EngineWithPrivateStateHandlers[User]) AddMiddleware(
	middleware func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)) *EngineWithPrivateStateHandlers[User] {

	e.m.Lock()
	defer e.m.Unlock()

	e.middlewares = append(e.middlewares, middleware)

	return e
}

// AddInlineMenu adds an inline state handler
func (e *EngineWithPrivateStateHandlers[User]) AddInlineMenu(
	name string, handler *InlineMenu[User]) *EngineWithPrivateStateHandlers[User] {

	e.m.Lock()
	defer e.m.Unlock()

	handler.callbackPrefix = name
	if handler.inlineActionBuilder != nil {
		handler.inlineActionBuilder.inlineMenu = name
	}

	e.inlineMenus[name] = handler

	return e
}

// WithLanguageConfig adds a language config to the engine
func (e *EngineWithPrivateStateHandlers[User]) WithLanguageConfig(
	cfg *LanguageConfig) *EngineWithPrivateStateHandlers[User] {

	e.languageConfig = cfg

	if cfg.changeLanguageState == "" {
		return e
	}

	menu := NewStaticMenu[User]()

	menu.AddMiddleware(func(bot *tgbotapi.TelegramBot, update *StateUpdate[User]) (string, bool) {
		actions := NewActionBuilder()

		for i := range cfg.languages.localizers {
			lang := cfg.languages.localizers[i]

			btnText, _ := lang.Get(fmt.Sprintf("%s.Button", cfg.changeLanguageState))
			if btnText == "" {
				btnText = lang.tag
			}

			actions.AddCustomButton(NewChooseLanguageButton[User](btnText, e, update, cfg, &lang, bot))
		}

		menu.WithStaticActionBuilder(actions)

		return "", true
	})

	text := ""
	for _, lang := range cfg.languages.localizers {
		txt, _ := lang.Get(fmt.Sprintf("%s.Text", cfg.changeLanguageState))
		if txt == "" {
			txt = cfg.changeLanguageState
		}

		text += fmt.Sprintf("%s\n", txt)
	}

	menu.ReplyWithText(text)

	return e.AddStaticMenu(cfg.changeLanguageState, menu)
}

func (e *EngineWithPrivateStateHandlers[User]) Process(client *tgbotapi.TelegramBot, update tgbotapi.Update) {
	user, userState, err := e.processUserState(update)
	if err != nil {
		e.onErr(client, update, err)
		return
	}

	su := &StateUpdate[User]{
		storage:    &sync.Map{},
		State:      userState,
		User:       user,
		Update:     update,
		IsSwitched: false,
	}

	from := update.From()

	if from == nil {
		return
	}

	var lang *Language

	if e.languageConfig != nil {
		userLanguage, err := e.languageConfig.repo.GetUserLanguage(from.Id)
		if err != nil {
			if err == UserLanguageNotFoundErr && e.languageConfig.forceChooseLanguage {
				if userState != e.languageConfig.changeLanguageState {
					err = e.switchState(
						from.Id, e.languageConfig.changeLanguageState, client, su)
					if err != nil {
						e.onErr(client, update, err)
					}
					return
				}
			} else {
				e.onErr(client, update, err)
				return
			}
		}

		lang = e.languageConfig.languages.GetByTag(userLanguage)
	}

	su.language = lang

	for _, f := range e.middlewares {
		if target, ok := f(client, su); !ok {
			if target != "" {
				if err := e.switchState(from.Id, target, client, su); err != nil {
					e.onErr(client, update, err)
				}
			}
			return
		}
	}

	if update.Message != nil {
		if handler := e.staticMenus[userState]; handler != nil {
			e.processStaticHandler(from.Id, handler, client, su)
			return
		}
	}

	if update.CallbackQuery != nil {
		e.processCallbackQuery(client, su)
		return
	}
}

// AddCallbackQueryHandler adds a callback query handler
func (e *EngineWithPrivateStateHandlers[User]) AddCallbackQueryHandler(
	data string,
	fn func(*tgbotapi.TelegramBot, *StateUpdate[User], ...string) error) *EngineWithPrivateStateHandlers[User] {

	e.m.Lock()
	defer e.m.Unlock()

	e.callbackQueryHandlers[data] = fn

	return e
}

func (e *EngineWithPrivateStateHandlers[User]) SwitchState(
	userID int64, client *tgbotapi.TelegramBot, update *StateUpdate[User], state string) error {

	return e.switchState(userID, state, client, update)
}

func (e *EngineWithPrivateStateHandlers[User]) SwitchUserState(
	client *tgbotapi.TelegramBot, userID int64, state string) error {

	user, lang, err := e.userInfo(userID)
	if err != nil {
		return err
	}

	return e.switchState(userID, state, client, &StateUpdate[User]{
		storage:    &sync.Map{},
		State:      state,
		User:       user,
		language:   lang,
		IsSwitched: true,
	})
}

func (e *EngineWithPrivateStateHandlers[User]) SendInlineMenu(
	client *tgbotapi.TelegramBot, update *StateUpdate[User], menu string, shouldEdit bool) error {

	return e.processInlineHandler(menu, client, update, shouldEdit)
}

// getCallbackQueryHandler returns a callback query handler by data
func (e *EngineWithPrivateStateHandlers[User]) getCallbackQueryHandler(
	data string) func(*tgbotapi.TelegramBot, *StateUpdate[User], ...string) error {

	e.m.Lock()
	defer e.m.Unlock()

	if handler, ok := e.callbackQueryHandlers[data]; ok {
		return handler
	}

	return nil
}

func (e *EngineWithPrivateStateHandlers[User]) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil && chat.Type == "private" {
		return true
	}

	return false
}

func (e *EngineWithPrivateStateHandlers[User]) processCallbackQuery(
	client *tgbotapi.TelegramBot, update *StateUpdate[User]) {

	if update.Update.CallbackQuery.Data == "" {
		return
	}

	data := strings.Split(update.Update.CallbackQuery.Data, ":")

	menu := data[0]

	if inlineMenu, ok := e.inlineMenus[menu]; !ok {
		if callbackHandler := e.getCallbackQueryHandler(data[0]); callbackHandler != nil {
			err := callbackHandler(client, update, data[1:]...)
			if err != nil {
				e.onErr(client, update.Update, err)
			}
		} else {
			e.onErr(client, update.Update, errors.New("callback query handler not found: "+data[0]))
		}
		return
	} else {
		for _, f := range inlineMenu.middlewares {
			if !f(client, update) {
				return
			}
		}

		if err := e.processInlineCallbackHandler(client, update, inlineMenu, data[1:]); err != nil {
			e.onErr(client, update.Update, errors.New("error processing inline menu: "+err.Error()))
		}

		return
	}
}

func (e *EngineWithPrivateStateHandlers[User]) processStaticHandler(
	userID int64, handler *StaticMenu[User], client *tgbotapi.TelegramBot, update *StateUpdate[User]) {

	for _, middleware := range handler.middlewares {
		if target, ok := middleware(client, update); !ok {
			if target != "" {
				if err := e.switchState(userID, target, client, update); err != nil {
					e.onErr(client, update.Update, err)
				}
			}

			return
		}
	}

	if update.Update.Message != nil && update.Update.Message.Text != "" {
		if !update.IsSwitched {
			buttonText := update.Update.Message.Text

			if update.language != nil && handler.actionBuilder != nil {
				if languageValueKeys := handler.actionBuilder.languageValueButtonKeys(update.language); languageValueKeys != nil {
					if languageValueKey := languageValueKeys[buttonText]; languageValueKey != "" {
						buttonText = languageValueKey
					}
				}
			}

			if handler.actionBuilder != nil {
				if buttonAction := handler.actionBuilder.getButtonByButton(buttonText); buttonAction != nil {
					var err error

					shouldStop := true

					switch buttonAction.Kind() {
					case ActionKindText:
						text := buttonAction.Result()
						_, err = client.Send(client.Message().SetText(text).SetChatId(userID))
						if err != nil {
							err = fmt.Errorf("error_sending_message_to_user: %d, %w", userID, err)
						}
					case ActionKindState:
						nextState := buttonAction.Result()
						if err := e.switchState(userID, nextState, client, update); err != nil {
							err = fmt.Errorf("error_switching_state: %d, %w", userID, err)
						}
					case ActionKindInlineMenu:
						inlineMenu := buttonAction.Result()
						err = e.processInlineHandler(inlineMenu, client, update, false)
						if err != nil {
							err = fmt.Errorf("error_switching_inline_menu: %d, %w", userID, err)
						}
					case ActionKindRaw:
						shouldStop = false
						// do nothing for raw action, as it is only used to act like a button and may be handled in a
						// dynamic handler
						break
					default:
						err = fmt.Errorf("unknown_action_kind: %s", buttonAction.Kind())
					}

					if err != nil {
						e.onErr(client, update.Update, err)
						return
					}

					if shouldStop {
						return
					}
				}
			}

			if handler.dynamicHandlers != nil && handler.dynamicHandlers.textHandler != nil {
				if nextState, next := handler.dynamicHandlers.textHandler(client, update); nextState != "" {
					err := e.switchState(userID, nextState, client, update)
					if err != nil {
						e.onErr(client, update.Update, err)
					}
					return
				} else if !next {
					return
				}
			}
		}
	}

	var replyMarkup *structs.ReplyKeyboardMarkup

	if handler.actionBuilder != nil {
		replyMarkup = handler.actionBuilder.buildButtons(update.language)
	}

	if replyText := handler.getReplyText(); replyText != "" {
		cfg := client.Message().SetText(replyText).SetChatId(userID)
		if replyMarkup != nil {
			cfg = cfg.SetReplyMarkup(replyMarkup)
		}
		_, err := client.Send(cfg)
		if err != nil {
			e.onErr(client, update.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", userID, err))
			return
		}
	}

	if replyLanguageKey := handler.getReplyTextLanguageKey(); replyLanguageKey != "" {
		var txt string
		if update.language == nil {
			txt = replyLanguageKey
		} else {
			result, err := update.language.Get(replyLanguageKey)
			if err == nil {
				txt = result
			}
		}

		cfg := client.Message().SetText(txt).SetChatId(userID)
		if replyMarkup != nil {
			cfg = cfg.SetReplyMarkup(replyMarkup)
		}
		_, err := client.Send(cfg)
		if err != nil {
			e.onErr(client, update.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", userID, err))
			return
		}
	}

	if replyWithFunc := handler.getReplyWithFunc(); replyWithFunc != nil {
		replyWithFunc(client, update)
	}
}

func (e *EngineWithPrivateStateHandlers[User]) processInlineHandler(
	menuName string, client *tgbotapi.TelegramBot, update *StateUpdate[User], edit bool) error {

	menu, ok := e.inlineMenus[menuName]
	if !ok {
		return fmt.Errorf("inline_menu_not_found: %s", menuName)
	}

	from := update.Update.From()

	if middlewares := menu.getMiddlewares(); len(middlewares) > 0 {
		for _, middleware := range middlewares {
			if canProceed := middleware(client, update); !canProceed {
				return nil
			}
		}
	}

	if menu.inlineActionBuilder == nil {
		return fmt.Errorf("inline_menu_action_builder_not_set: %s", menuName)
	}

	markup := menu.inlineActionBuilder.buildButtons(update.language)

	if replyText := menu.getReplyText(); replyText != "" {
		var cfg tgbotapi.Config

		if edit {
			cfg = client.EditMessageText().SetText(menu.replyText).
				SetChatId(from.Id).
				SetMessageId(update.Update.CallbackQuery.Message.MessageId).
				SetReplyMarkup(markup)
		} else {
			cfg = client.Message().SetText(menu.replyText).SetChatId(from.Id).SetReplyMarkup(markup)
		}

		_, err := client.Send(cfg)
		if err != nil {
			return fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err)
		}
	}

	return nil
}

func (e *EngineWithPrivateStateHandlers[User]) switchState(
	userID int64, nextState string, client *tgbotapi.TelegramBot, stateUpdate *StateUpdate[User]) error {

	if handler := e.staticMenus[nextState]; handler != nil {
		if err := e.userRepository.SetState(userID, nextState); err != nil {
			return fmt.Errorf("error_setting_user_state: %d, %w", userID, err)
		}

		stateUpdate.State = nextState
		stateUpdate.IsSwitched = true

		e.processStaticHandler(userID, handler, client, stateUpdate)

		return nil
	}

	return fmt.Errorf("no_handler_for_state: %s", nextState)
}

func (e *EngineWithPrivateStateHandlers[User]) processUserState(update tgbotapi.Update) (User, string, error) {
	from := update.From()

	user, err := e.userRepository.Find(from.Id)
	if err != nil {
		user, err = e.userRepository.Store(from)
		if err != nil {
			return *new(User), "", fmt.Errorf("store_user: %w", err)
		}
	}

	if e.defaultStateName == "" {
		return *new(User), "", fmt.Errorf("empty_default_state_name")
	}

	userState, err := e.userRepository.GetState(from.Id)
	if err != nil || userState == "" {
		userState = e.defaultStateName
		err = e.userRepository.SetState(from.Id, userState)
		if err != nil {
			return *new(User), "", fmt.Errorf("store_user_state: %w", err)
		}
	}

	return user, userState, nil
}

func (e *EngineWithPrivateStateHandlers[User]) userInfo(userID int64) (User, *Language, error) {
	user, err := e.userRepository.Find(userID)
	if err != nil {
		return *new(User), nil, fmt.Errorf("find_user: %w", err)
	}

	var lang *Language

	if e.languageConfig != nil {
		userLanguage, _ := e.languageConfig.repo.GetUserLanguage(userID)
		if userLanguage != "" {
			lang = e.languageConfig.languages.GetByTag(userLanguage)
		}
	}

	return user, lang, nil

}

// getHandlerByAction returns inline menu by action
func (e *EngineWithPrivateStateHandlers[User]) getHandlerByAction(
	client *tgbotapi.TelegramBot, update *StateUpdate[User], action string) (InlineAction, error) {

	for _, menu := range e.inlineMenus {
		for _, f := range menu.middlewares {
			if !f(client, update) {
				return nil, nil
			}
		}

		if menu.inlineActionBuilder == nil {
			continue
		}

		handlers := menu.inlineActionBuilder.getByCallbackActionData()
		if handlers == nil {
			continue
		}

		if handler, ok := handlers[action]; ok {
			return handler, nil
		}
	}

	return nil, fmt.Errorf("handler_for_action_not_found: %s", action)
}

// getHandlerByAction returns inline menu by action
func (e *EngineWithPrivateStateHandlers[User]) processInlineCallbackHandler(
	client *tgbotapi.TelegramBot, update *StateUpdate[User], menu *InlineMenu[User], data []string) error {

	for _, f := range menu.middlewares {
		if !f(client, update) {
			return nil
		}
	}

	actionHandlers := menu.inlineActionBuilder.getByCallbackActionData()

	if actionHandlers == nil {
		return fmt.Errorf("inline_menu_action_builder_not_set: %s", menu.callbackPrefix)
	}

	if handler, ok := actionHandlers[data[0]]; !ok {
		return fmt.Errorf("handler_for_action_not_found: %s", data[0])
	} else {
		switch btn := handler.(type) {
		case inlineAlertButton:
			cfg := client.AnswerCallbackQuery().
				SetCallbackQueryId(update.Update.CallbackQuery.Id).
				SetText(btn.text).
				SetShowAlert(btn.showAlert)
			_, err := client.Send(cfg)

			return err
		case inlineStateButton:
			// TODO: Implement Switch With Edit Message

			return e.switchState(update.Update.From().Id, btn.state, client, update)
		case inlineInlineMenuButton:
			return e.processInlineHandler(btn.menu, client, update, btn.edit)
		case inlineCallbackButton:
			if callbackHandler := e.getCallbackQueryHandler(data[0]); callbackHandler != nil {
				callbackHandler(client, update, data[1:]...)
				return nil
			}

			return errors.New("callback query handler not found")
		}
	}

	return errors.New("processor_for_action_not_found")
}
