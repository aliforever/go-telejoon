package telejoon

import (
	"context"
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

	languageConfig *LanguageConfig
}

func WithPrivateStateHandlers[User any](
	userRepo UserRepository[User], defaultState string, opts ...*Options) *EngineWithPrivateStateHandlers[User] {

	return &EngineWithPrivateStateHandlers[User]{
		engine: engine[User, any, any, any]{
			opts: opts,
		},
		userRepository:   userRepo,
		defaultStateName: defaultState,
		staticMenus:      map[string]*StaticMenu[User]{},
		inlineMenus:      map[string]*InlineMenu[User]{},
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

func (e *EngineWithPrivateStateHandlers[User]) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil && chat.Type == "private" {
		return true
	}

	return false
}

func (e *EngineWithPrivateStateHandlers[User]) process(client *tgbotapi.TelegramBot, update tgbotapi.Update) {
	user, userState, err := e.processUserState(update)
	if err != nil {
		e.onErr(client, update, err)
		return
	}

	su := &StateUpdate[User]{
		context:    context.WithValue(context.Background(), "storage", &sync.Map{}),
		State:      userState,
		User:       user,
		Update:     update,
		IsSwitched: false,
	}

	var lang *Language

	if e.languageConfig != nil {
		userLanguage, err := e.languageConfig.repo.GetUserLanguage(update.From().Id)
		if err != nil {
			if err == UserLanguageNotFoundErr && e.languageConfig.forceChooseLanguage {
				if userState != e.languageConfig.changeLanguageState {
					err = e.switchState(
						e.languageConfig.changeLanguageState, client, su)
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

		lang = e.languageConfig.languages.getByTag(userLanguage)
	}

	su.language = lang

	for _, f := range e.middlewares {
		if nextState, ok := f(client, su); !ok {
			if nextState != "" {
				if err := e.switchState(nextState, client, su); err != nil {
					e.onErr(client, update, err)
				}
			}
			return
		}
	}

	if update.Message != nil {
		if handler := e.staticMenus[userState]; handler != nil {
			e.processStaticHandler(handler, client, su)
			return
		}
	}

	if update.CallbackQuery != nil {
		e.processCallbackQuery(client, su)
		return
	}
}

func (e *EngineWithPrivateStateHandlers[User]) processCallbackQuery(
	client *tgbotapi.TelegramBot, update *StateUpdate[User]) {

	if update.Update.CallbackQuery.Data == "" {
		return
	}

	data := strings.Split(update.Update.CallbackQuery.Data, ":")

	command := data[0]

	for _, menu := range e.inlineMenus {
		if dba, ok := menu.buttonAlerts[command]; ok {
			_, _ = client.Send(client.AnswerCallbackQuery().
				SetCallbackQueryId(update.Update.CallbackQuery.Id).
				SetText(dba.text).
				SetShowAlert(dba.showAlert))
			return
		}

		if inlineMenu, ok := menu.buttonInlineMenus[command]; ok {
			err := e.processInlineHandler(inlineMenu.data, client, update, inlineMenu.edit)
			if err != nil {
				e.onErr(client, update.Update, err)
			}
			return
		}
	}
}

func (e *EngineWithPrivateStateHandlers[User]) processStaticHandler(
	handler *StaticMenu[User], client *tgbotapi.TelegramBot, update *StateUpdate[User]) {

	from := update.Update.From()

	for _, middleware := range handler.middlewares {
		if nextState, ok := middleware(client, update); !ok {
			if nextState != "" {
				if err := e.userRepository.SetState(from.Id, nextState); err != nil {
					e.onErr(client, update.Update,
						fmt.Errorf("error_setting_user_state: %d, %w", from.Id, err))
					return
				}

				e.processStaticHandler(e.staticMenus[nextState], client, &StateUpdate[User]{
					context:    update.context,
					State:      nextState,
					User:       update.User,
					Update:     update.Update,
					language:   update.language,
					IsSwitched: true,
				})
			}

			return
		}
	}

	if update.Update.Message != nil && update.Update.Message.Text != "" {
		if !update.IsSwitched {
			buttonText := update.Update.Message.Text

			if update.language != nil && handler.staticActionBuilder != nil {
				if languageValueKeys := handler.staticActionBuilder.languageValueButtonKeys(update.language); languageValueKeys != nil {
					if languageValueKey := languageValueKeys[buttonText]; languageValueKey != "" {
						buttonText = languageValueKey
					}
				}
			}

			if handler.staticActionBuilder != nil {
				if buttonAction := handler.staticActionBuilder.getButtonByButton(buttonText); buttonAction != nil {
					var err error

					switch buttonAction.Kind() {
					case ActionKindText:
						text := buttonAction.Result()
						_, err = client.Send(client.Message().SetText(text).SetChatId(from.Id))
						if err != nil {
							err = fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err)
						}
					case ActionKindState:
						nextState := buttonAction.Result()
						if err := e.switchState(nextState, client, update); err != nil {
							err = fmt.Errorf("error_switching_state: %d, %w", from.Id, err)
						}
					case ActionKindInlineMenu:
						inlineMenu := buttonAction.Result()
						err = e.processInlineHandler(inlineMenu, client, update, false)
						if err != nil {
							err = fmt.Errorf("error_switching_inline_menu: %d, %w", from.Id, err)
						}
					default:
						err = fmt.Errorf("unknown_action_kind: %s", buttonAction.Kind())
					}

					if err != nil {
						e.onErr(client, update.Update, err)
						return
					}

					return
				}
			}
		}
	}

	var replyMarkup *structs.ReplyKeyboardMarkup

	if handler.staticActionBuilder != nil {
		replyMarkup = handler.staticActionBuilder.buildButtons(update.language)
	}

	if replyText := handler.getReplyText(); replyText != "" {
		cfg := client.Message().SetText(replyText).SetChatId(from.Id)
		if replyMarkup != nil {
			cfg = cfg.SetReplyMarkup(replyMarkup)
		}
		_, err := client.Send(cfg)
		if err != nil {
			e.onErr(client, update.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err))
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

		cfg := client.Message().SetText(txt).SetChatId(from.Id)
		if replyMarkup != nil {
			cfg = cfg.SetReplyMarkup(replyMarkup)
		}
		_, err := client.Send(cfg)
		if err != nil {
			e.onErr(client, update.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err))
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

	markup := menu.getInlineKeyboardMarkup(update.language)

	if menu.replyText != "" {
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
	nextState string, client *tgbotapi.TelegramBot, stateUpdate *StateUpdate[User]) error {

	from := stateUpdate.Update.From()

	if handler := e.staticMenus[nextState]; handler != nil {
		if err := e.userRepository.SetState(from.Id, nextState); err != nil {
			return fmt.Errorf("error_setting_user_state: %d, %w", from.Id, err)
		}

		e.processStaticHandler(handler, client, &StateUpdate[User]{
			context:    stateUpdate.context,
			State:      nextState,
			language:   stateUpdate.language,
			User:       stateUpdate.User,
			Update:     stateUpdate.Update,
			IsSwitched: true,
		})

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
