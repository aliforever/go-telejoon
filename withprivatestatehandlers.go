package telejoon

import (
	"context"
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
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

	for _, lang := range cfg.languages.localizers {
		btnText, _ := lang.Get(fmt.Sprintf("%s.Button", cfg.changeLanguageState))
		if btnText == "" {
			btnText = lang.tag
		}

		menu.AddButtonFunc(btnText,
			func(bot *tgbotapi.TelegramBot, update *StateUpdate[User]) string {
				err := cfg.repo.SetUserLanguage(update.Update.From().Id, lang.tag)
				if err != nil {
					e.onErr(bot, update.Update, err)
					return ""
				}

				return e.defaultStateName
			})
	}

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

	var lang Language

	if e.languageConfig != nil {
		userLanguage, err := e.languageConfig.repo.GetUserLanguage(update.From().Id)
		if err != nil {
			if err == UserLanguageNotFoundErr && e.languageConfig.forceChooseLanguage {
				if userState != e.languageConfig.changeLanguageState {
					err = e.switchState(e.languageConfig.changeLanguageState, client, context.Background(), user, update)
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

	su := &StateUpdate[User]{
		context:    context.Background(),
		State:      userState,
		User:       user,
		Language:   lang,
		Update:     update,
		IsSwitched: false,
	}

	for _, f := range e.middlewares {
		if nextState, ok := f(client, su); !ok {
			if nextState != "" {
				if err := e.switchState(nextState, client, su.context, user, update); err != nil {
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
		// client.Send(client.Message().SetText(string(j)).SetChatId(update.From().Id))
		if update.CallbackQuery.Data == "" {
			return
		}

		data := strings.Split(update.CallbackQuery.Data, ":")

		command := data[0]

		for _, menu := range e.inlineMenus {
			if dba, ok := menu.buttonAlerts[command]; ok {
				_, _ = client.Send(client.AnswerCallbackQuery().
					SetCallbackQueryId(update.CallbackQuery.Id).
					SetText(dba.text).
					SetShowAlert(dba.showAlert))
				return
			}

			if inlineMenu, ok := menu.buttonInlineMenus[command]; ok {
				if inlineMenuHandler, ok := e.inlineMenus[inlineMenu.data]; ok {
					e.processInlineHandler(inlineMenuHandler, client, su, inlineMenu.edit)
					return
				}
			}
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
					IsSwitched: true,
				})
			}

			return
		}
	}

	replyMarkup := handler.buildButtonKeyboard()

	if update.Update.Message != nil && update.Update.Message.Text != "" {
		if !update.IsSwitched {
			if response := handler.getReplyTextForButton(update.Update.Message.Text); response != "" {
				_, err := client.Send(client.Message().SetText(response).SetChatId(from.Id))
				if err != nil {
					e.onErr(client, update.Update,
						fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err))
					return
				}
				return
			}

			if nextState := handler.getStateForButton(update.Update.Message.Text); nextState != "" {
				if err := e.switchState(nextState, client, update.context, update.User, update.Update); err != nil {
					e.onErr(client, update.Update, err)
				}
				return
			}

			if fn := handler.getFuncForButton(update.Update.Message.Text); fn != nil {
				if nextState := fn(client, update); nextState != "" {
					if err := e.switchState(nextState, client, update.context, update.User, update.Update); err != nil {
						e.onErr(client, update.Update, err)
					}
				}
				return
			}

			if inlineMenu := handler.getInlineMenuForButton(update.Update.Message.Text); inlineMenu != "" {
				if menu, ok := e.inlineMenus[inlineMenu]; !ok {
					e.onErr(client, update.Update, fmt.Errorf("inline_menu_not_found: %s", inlineMenu))
				} else {
					e.processInlineHandler(menu, client, update, false)
				}
				return
			}
		}
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

	if replyWithFunc := handler.getReplyWithFunc(); replyWithFunc != nil {
		replyWithFunc(client, update)
	}
}

func (e *EngineWithPrivateStateHandlers[User]) processInlineHandler(
	menu *InlineMenu[User], client *tgbotapi.TelegramBot, update *StateUpdate[User], edit bool) {

	from := update.Update.From()

	markup := menu.getInlineKeyboardMarkup()

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
			e.onErr(client, update.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err))
			return
		}
	}
}

func (e *EngineWithPrivateStateHandlers[User]) switchState(
	nextState string, client *tgbotapi.TelegramBot, ctx context.Context, user User, update tgbotapi.Update) error {

	from := update.From()

	if handler := e.staticMenus[nextState]; handler != nil {
		if err := e.userRepository.SetState(from.Id, nextState); err != nil {
			return fmt.Errorf("error_setting_user_state: %d, %w", from.Id, err)
		}
		e.processStaticHandler(handler, client, &StateUpdate[User]{
			context:    ctx,
			State:      nextState,
			User:       user,
			Update:     update,
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
