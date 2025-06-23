package telejoon

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
)

type EngineWithPrivateStateHandlers struct {
	engine

	userRepository UserRepository

	m sync.Mutex

	panicHandler PanicHandler

	middlewares []UpdateHandler

	defaultStateName string

	staticMenus map[string]*StaticMenu

	inlineMenus map[string]*InlineMenu

	callbackQueryHandlers map[string]func(
		client *tgbotapi.TelegramBot, update *StateUpdate, args ...string) (SwitchAction, error)

	languageConfig *LanguageConfig
}

func WithPrivateStateHandlers(
	userRepo UserRepository, defaultState string, opts ...*Options) *EngineWithPrivateStateHandlers {

	return &EngineWithPrivateStateHandlers{
		engine: engine{
			opts: opts,
		},
		userRepository:   userRepo,
		defaultStateName: defaultState,
		staticMenus:      map[string]*StaticMenu{},
		inlineMenus:      map[string]*InlineMenu{},
		callbackQueryHandlers: map[string]func(
			*tgbotapi.TelegramBot, *StateUpdate, ...string) (SwitchAction, error){},
	}
}

// AddStaticMenu adds a static state Handler
func (e *EngineWithPrivateStateHandlers) AddStaticMenu(
	state string,
	handler *StaticMenu,
) *EngineWithPrivateStateHandlers {

	e.m.Lock()
	defer e.m.Unlock()

	e.staticMenus[state] = handler

	return e
}

func (e *EngineWithPrivateStateHandlers) WithPanicHandler(
	handler PanicHandler,
) *EngineWithPrivateStateHandlers {

	e.m.Lock()
	defer e.m.Unlock()

	e.panicHandler = handler

	return e
}

func (e *EngineWithPrivateStateHandlers) AddMiddleware(
	middleware UpdateHandler,
) *EngineWithPrivateStateHandlers {

	e.m.Lock()
	defer e.m.Unlock()

	e.middlewares = append(e.middlewares, middleware)

	return e
}

// AddInlineMenu adds an inline state Handler
func (e *EngineWithPrivateStateHandlers) AddInlineMenu(
	name string,
	handler *InlineMenu,
) *EngineWithPrivateStateHandlers {

	e.m.Lock()
	defer e.m.Unlock()

	handler.callbackPrefix = name

	e.inlineMenus[name] = handler

	return e
}

// WithLanguageConfig adds a language config to the engine
func (e *EngineWithPrivateStateHandlers) WithLanguageConfig(cfg *LanguageConfig) *EngineWithPrivateStateHandlers {
	e.languageConfig = cfg

	if cfg.changeLanguageState == "" {
		return e
	}

	text := ""

	for _, lang := range cfg.languages.localizers {
		txt, _ := lang.Get(fmt.Sprintf("%s.Text", cfg.changeLanguageState))
		if txt == "" {
			txt = cfg.changeLanguageState
		}

		text += fmt.Sprintf("%s\n", txt)
	}

	deferredActionBuilder := NewDeferredActionBuilder(func(update *StateUpdate) *ActionBuilder {
		actions := NewStaticActionBuilder()

		for i := range cfg.languages.localizers {
			lang := cfg.languages.localizers[i]

			btnText, _ := lang.Get(fmt.Sprintf("%s.Button", cfg.changeLanguageState))
			if btnText == "" {
				btnText = lang.tag
			}

			actions.AddRawButton(NewStaticText(btnText))
		}

		return actions
	})

	deferredDynamicTextBuilder := NewDynamicHandlerText(func(
		client *tgbotapi.TelegramBot,
		update *StateUpdate,
	) (SwitchAction, ShouldPass) {

		for i := range cfg.languages.localizers {
			lang := cfg.languages.localizers[i]

			btnText, _ := lang.Get(fmt.Sprintf("%s.Button", cfg.changeLanguageState))
			if btnText == "" {
				btnText = lang.tag
			}

			if update.Update.Message.Text == btnText {
				err := e.languageConfig.repo.SetUserLanguage(update.Update.From().Id, lang.tag)
				if err != nil {
					e.engine.onErr(client, update.Update, err)
					return nil, false
				}

				update.SetLanguage(&lang)

				return NewSwitchActionState(e.defaultStateName), false
			}
		}

		return nil, true
	})

	menu := NewStaticMenu(
		NewStaticText(text),
		deferredActionBuilder,
		deferredDynamicTextBuilder)

	return e.AddStaticMenu(cfg.changeLanguageState, menu)
}

func (e *EngineWithPrivateStateHandlers) Process(client *tgbotapi.TelegramBot, update tgbotapi.Update) {
	if e.panicHandler != nil {
		defer func() {
			if r := recover(); r != nil {
				e.panicHandler(client, update, r, string(debug.Stack()))
			}
		}()
	}

	userState, err := e.processUserState(update)
	if err != nil {
		e.onErr(client, update, err)
		return
	}

	su := &StateUpdate{
		storage:    &sync.Map{},
		State:      userState,
		Update:     update,
		IsSwitched: false,
	}

	from := update.From()

	if from == nil {
		j, _ := json.Marshal(update)
		e.onErr(client, update, fmt.Errorf("update.From() is nil: %s", string(j)))
		return
	}

	var lang *Language

	if e.languageConfig != nil {
		userLanguage, err := e.languageConfig.repo.GetUserLanguage(from.Id)
		if err != nil {
			if e.languageConfig.forceChooseLanguage {
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
		switchAction, pass := f.Handle(client, su)
		if err := e.processSwitchAction(switchAction, su, client); err != nil {
			e.onErr(client, update, err)
			return
		}

		if !pass {
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

// AddCallbackQueryHandler adds a callback query Handler
func (e *EngineWithPrivateStateHandlers) AddCallbackQueryHandler(
	data string,
	fn func(*tgbotapi.TelegramBot, *StateUpdate, ...string) (SwitchAction, error),
) *EngineWithPrivateStateHandlers {

	e.m.Lock()
	defer e.m.Unlock()

	e.callbackQueryHandlers[data] = fn

	return e
}

func (e *EngineWithPrivateStateHandlers) SwitchState(
	userID int64, client *tgbotapi.TelegramBot, update *StateUpdate, state string) error {

	return e.switchState(userID, state, client, update)
}

func (e *EngineWithPrivateStateHandlers) SwitchUserState(
	client *tgbotapi.TelegramBot, userID int64, state string) error {

	lang, err := e.userLanguage(userID)
	if err != nil {
		return err
	}

	return e.switchState(userID, state, client, &StateUpdate{
		storage:    &sync.Map{},
		State:      state,
		language:   lang,
		IsSwitched: true,
	})
}

func (e *EngineWithPrivateStateHandlers) SendInlineMenu(
	client *tgbotapi.TelegramBot, update *StateUpdate, menu string, shouldEdit bool) error {

	return e.processInlineHandler(menu, client, update, shouldEdit)
}

// getCallbackQueryHandler returns a callback query Handler by data
func (e *EngineWithPrivateStateHandlers) getCallbackQueryHandler(
	data string) func(*tgbotapi.TelegramBot, *StateUpdate, ...string) (SwitchAction, error) {

	e.m.Lock()
	defer e.m.Unlock()

	if handler, ok := e.callbackQueryHandlers[data]; ok {
		return handler
	}

	return nil
}

func (e *EngineWithPrivateStateHandlers) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil && chat.Type == "private" {
		return true
	}

	return false
}

func (e *EngineWithPrivateStateHandlers) processCallbackQuery(
	client *tgbotapi.TelegramBot,
	update *StateUpdate,
) {

	if update.Update.CallbackQuery.Data == "" {
		return
	}

	data := strings.Split(update.Update.CallbackQuery.Data, ":")

	menu := data[0]

	if inlineMenu, ok := e.inlineMenus[menu]; !ok {
		if callbackHandler := e.getCallbackQueryHandler(data[0]); callbackHandler != nil {
			switchAction, err := callbackHandler(client, update, data[1:]...)
			if err != nil {
				e.onErr(client, update.Update, err)
				return
			}

			if err := e.processSwitchAction(switchAction, update, client); err != nil {
				e.onErr(client, update.Update, err)
			}
		} else {
			e.onErr(client, update.Update, errors.New("callback query Handler not found: "+data[0]))
		}
		return
	} else {
		if err := e.processInlineCallbackHandler(client, update, inlineMenu, data[1:]); err != nil {
			e.onErr(client, update.Update, errors.New("error processing inline menu: "+err.Error()))
		}

		return
	}
}

func (e *EngineWithPrivateStateHandlers) processStaticHandler(
	userID int64,
	handler *StaticMenu,
	client *tgbotapi.TelegramBot,
	update *StateUpdate,
) {

	for _, middleware := range handler.middlewares {
		if middleware.UpdateHandler == nil {
			continue
		}

		switchAction, pass := middleware.Handle(client, update)

		if err := e.processSwitchAction(switchAction, update, client); err != nil {
			e.onErr(client, update.Update, err)
			return
		}

		if !pass {
			return
		}
	}

	actionBuilder := handler.processActionBuilder(update)

	if !update.IsSwitched {
		if update.Update.Message != nil && update.Update.Message.Text != "" {
			buttonText := update.Update.Message.Text

			if actionBuilder != nil {
				if buttonAction := actionBuilder.getButtonByButton(
					update,
					buttonText,
				); buttonAction != nil {
					var err error

					shouldStop := true

					switch a := buttonAction.(type) {
					case textButton:
						if _, err = client.
							Send(client.Message().SetText(a.text.String(update)).SetChatId(userID)); err != nil {
							err = fmt.Errorf("error_sending_message_to_user: %d, %w", userID, err)
						}
					case stateButton:
						if a.hook != nil {
							switchAction, pass := a.hook.Handle(client, update)
							if err = e.processSwitchAction(switchAction, update, client); err != nil {
								e.onErr(client, update.Update, err)
								return
							}

							if !pass {
								return
							}
						}

						if err = e.switchState(userID, a.state, client, update); err != nil {
							err = fmt.Errorf("error_switching_state: %d, %w", userID, err)
						}
					case inlineMenuButton:
						err = e.processInlineHandler(a.inlineMenu, client, update, false)
						if err != nil {
							err = fmt.Errorf("error_switching_inline_menu: %d, %w", userID, err)
						}
					case rawButton:
						shouldStop = false
						// do nothing for raw action, as it is only used to act like a button and may be handled in a
						// dynamic Handler
						break
					default:
						err = fmt.Errorf("unknown_action_kind: %+v", buttonAction)
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

			if handler.dynamicHandlers != nil && handler.dynamicHandlers[TextHandler] != nil {
				switchAction, pass := handler.dynamicHandlers[TextHandler].Handle(client, update)
				if err := e.processSwitchAction(switchAction, update, client); err != nil {
					e.onErr(client, update.Update, err)
					return
				}

				if !pass {
					return
				}
			}
		}

		if handler.dynamicHandlers != nil {
			handlerName := ""

			if update.Update.Message.Video != nil {
				handlerName = VideoHandler
			} else if update.Update.Message.Photo != nil {
				handlerName = PhotoHandler
			} else if update.Update.Message.Document != nil {
				handlerName = DocumentHandler
			} else if update.Update.Message.Voice != nil {
				handlerName = VoiceHandler
			} else if update.Update.Message.Audio != nil {
				handlerName = AudioHandler
			} else if update.Update.Message.Sticker != nil {
				handlerName = StickerHandler
			} else if update.Update.Message.Location != nil {
				handlerName = LocationHandler
			} else if update.Update.Message.Contact != nil {
				handlerName = ContactHandler
			} else if update.Update.Message.VideoNote != nil {
				handlerName = VideoNoteHandler
			}

			var targetHandler = handler.dynamicHandlers[DefaultHandler]

			if handlerName != "" {
				if availableHandler := handler.dynamicHandlers[handlerName]; availableHandler != nil {
					targetHandler = availableHandler
				}
			}

			if targetHandler != nil {
				switchAction, pass := targetHandler.Handle(client, update)

				if err := e.processSwitchAction(switchAction, update, client); err != nil {
					e.onErr(client, update.Update, err)
					return
				}

				if !pass {
					return
				}
			}
		}
	}

	var replyMarkup *structs.ReplyKeyboardMarkup

	if actionBuilder != nil {
		lang := update.Language()

		replyMarkup = actionBuilder.buildButtons(
			update,
			lang != nil && lang.rtl && e.languageConfig != nil && e.languageConfig.reverseButtonOrderInRowForRTL,
		)
	}

	if replyText := handler.processReplyText(update); replyText != "" {
		_, err := client.Send(client.Message().
			SetText(replyText).
			SetChatId(userID).
			SetReplyMarkup(replyMarkup))
		if err != nil {
			e.onErr(client, update.Update,
				fmt.Errorf("error_sending_message_to_user: %d, %w", userID, err))
			return
		}
	}
}

func (e *EngineWithPrivateStateHandlers) processInlineHandler(
	menuName string, client *tgbotapi.TelegramBot, update *StateUpdate, edit bool) error {

	menu, ok := e.inlineMenus[menuName]
	if !ok {
		return fmt.Errorf("inline_menu_not_found: %s", menuName)
	}

	from := update.Update.From()

	if middlewares := menu.getMiddlewares(); len(middlewares) > 0 {
		for _, middleware := range middlewares {
			switchAction, pass := middleware.Handle(client, update)
			if err := e.processSwitchAction(switchAction, update, client); err != nil {
				return err
			}

			if !pass {
				return nil
			}
		}
	}

	actionBuilder := menu.processActionBuilder(update)
	if actionBuilder == nil {
		return fmt.Errorf("inline_menu_action_builder_not_set: %s", menuName)
	}

	markup := actionBuilder.buildButtons(
		update,
		update.Language().rtl && e.languageConfig != nil && e.languageConfig.reverseButtonOrderInRowForRTL,
	)

	replyText := menu.processTextBuilder(update)
	if replyText == "" {
		return fmt.Errorf("inline_menu_reply_text_not_set: %s", menuName)
	}

	var cfg tgbotapi.Config

	if edit {
		cfg = client.EditMessageText().SetText(replyText).
			SetChatId(from.Id).
			SetMessageId(update.Update.CallbackQuery.Message.MessageId).
			SetReplyMarkup(markup)
	} else {
		cfg = client.Message().
			SetText(replyText).
			SetChatId(from.Id).
			SetReplyMarkup(markup)
	}

	_, err := client.Send(cfg)
	if err != nil {
		return fmt.Errorf("error_sending_message_to_user: %d, %w", from.Id, err)
	}

	return nil
}

func (e *EngineWithPrivateStateHandlers) switchState(
	userID int64, nextState string, client *tgbotapi.TelegramBot, stateUpdate *StateUpdate) error {

	if handler := e.staticMenus[nextState]; handler != nil {
		if err := e.userRepository.SetUserState(userID, nextState); err != nil {
			return fmt.Errorf("error_setting_user_state: %d, %w", userID, err)
		}

		stateUpdate.State = nextState
		stateUpdate.IsSwitched = true

		e.processStaticHandler(userID, handler, client, stateUpdate)

		return nil
	}

	return fmt.Errorf("no_handler_for_state: %s", nextState)
}

func (e *EngineWithPrivateStateHandlers) processUserState(update tgbotapi.Update) (string, error) {
	from := update.From()

	if from == nil {
		return "", errors.New("empty_from")
	}

	err := e.userRepository.UpsertUser(from)
	if err != nil {
		return "", fmt.Errorf("cant_store_user: %s", err)
	}

	if e.defaultStateName == "" {
		return "", fmt.Errorf("empty_default_state_name")
	}

	userState, err := e.userRepository.GetUserState(from.Id)
	if err != nil || userState == "" {
		userState = e.defaultStateName
		err = e.userRepository.SetUserState(from.Id, userState)
		if err != nil {
			return "", fmt.Errorf("store_user_state: %w", err)
		}
	}

	return userState, nil
}

func (e *EngineWithPrivateStateHandlers) userLanguage(userID int64) (*Language, error) {
	var lang *Language

	if e.languageConfig != nil {
		userLanguage, _ := e.languageConfig.repo.GetUserLanguage(userID)
		if userLanguage != "" {
			lang = e.languageConfig.languages.GetByTag(userLanguage)
		}
	}

	return lang, nil

}

// getHandlerByAction returns inline menu by action
func (e *EngineWithPrivateStateHandlers) processInlineCallbackHandler(
	client *tgbotapi.TelegramBot, update *StateUpdate, menu *InlineMenu, data []string) error {

	if middlewares := menu.getMiddlewares(); len(middlewares) > 0 {
		for _, middleware := range middlewares {
			switchAction, pass := middleware.Handle(client, update)
			if err := e.processSwitchAction(switchAction, update, client); err != nil {
				return err
			}

			if !pass {
				return nil
			}
		}
	}

	menuActionBuilder := menu.processActionBuilder(update)
	if menuActionBuilder == nil {
		return fmt.Errorf("inline_menu_action_builder_not_set: %s", menu.callbackPrefix)
	}

	actionHandlers := menuActionBuilder.getByCallbackActionData(update)
	if actionHandlers == nil {
		return fmt.Errorf("inline_menu_action_data_not_found: %s", menu.callbackPrefix)
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
			// TODO: Implement Switch With edit Message

			return e.switchState(update.Update.From().Id, btn.state, client, update)
		case inlineInlineMenuButton:
			return e.processInlineHandler(btn.menu, client, update, btn.edit)
		case inlineCallbackButton:
			if btn.handler != nil {
				switchAction, err := btn.handler(client, update, data[1:]...)
				if err != nil {
					return err
				}

				return e.processSwitchAction(switchAction, update, client)
			}

			return errors.New("callback query Handler not found")
		}
	}

	return errors.New("processor_for_action_not_found")
}

func (e *EngineWithPrivateStateHandlers) processSwitchAction(
	action SwitchAction,
	update *StateUpdate,
	client *tgbotapi.TelegramBot,
) error {

	if action == nil {
		return nil
	}

	switch sa := action.(type) {
	case *SwitchActionState:
		return e.switchState(update.Update.From().Id, action.target(), client, update)
	case *SwitchActionInlineMenu:
		return e.processInlineHandler(action.target(), client, update, sa.edit)
	}

	return errors.New("unknown switch action")
}
