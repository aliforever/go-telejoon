package telejoon

import (
	"context"
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type engineWithPrivateStateHandlers[User any] struct {
	engine[User, any, any, any]

	userRepository UserRepository[User]

	m sync.Mutex

	middlewares []func(client *tgbotapi.TelegramBot, update *StateUpdate[User]) (string, bool)

	defaultStateName string

	staticHandlers map[string]*StaticStateHandler[User]
}

func WithPrivateStateHandlers[User UserI[User]](
	userRepo UserRepository[User], defaultState string, opts ...*Options) *engineWithPrivateStateHandlers[User] {

	return &engineWithPrivateStateHandlers[User]{
		engine: engine[User, any, any, any]{
			opts: opts,
		},
		userRepository:   userRepo,
		defaultStateName: defaultState,
		staticHandlers:   map[string]*StaticStateHandler[User]{},
	}
}

// AddStaticHandler adds a static state handler
func (e *engineWithPrivateStateHandlers[User]) AddStaticHandler(
	state string, handler *StaticStateHandler[User]) *engineWithPrivateStateHandlers[User] {

	e.m.Lock()
	defer e.m.Unlock()

	e.staticHandlers[state] = handler

	return e
}

func (e *engineWithPrivateStateHandlers[User]) AddMiddleware(
	middleware func(*tgbotapi.TelegramBot, *StateUpdate[User]) (string, bool)) *engineWithPrivateStateHandlers[User] {

	e.m.Lock()
	defer e.m.Unlock()

	e.middlewares = append(e.middlewares, middleware)

	return e
}

func (e *engineWithPrivateStateHandlers[User]) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil && chat.Type == "private" {
		return true
	}

	return false
}

func (e *engineWithPrivateStateHandlers[User]) process(client *tgbotapi.TelegramBot, update tgbotapi.Update) {
	user, userState, err := e.processUserState(update)
	if err != nil {
		e.onErr(client, update, err)
		return
	}

	su := &StateUpdate[User]{
		context:    context.Background(),
		State:      userState,
		User:       user,
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

	if handler := e.staticHandlers[userState]; handler != nil {
		e.processStaticHandler(handler, client, su)
		return
	}
}

func (e *engineWithPrivateStateHandlers[User]) processStaticHandler(
	handler *StaticStateHandler[User], client *tgbotapi.TelegramBot, update *StateUpdate[User]) {

	from := update.Update.From()

	for _, middleware := range handler.middlewares {
		if nextState, ok := middleware(client, update); !ok {
			if nextState != "" {
				if err := e.userRepository.SetState(from.Id, nextState); err != nil {
					e.onErr(client, update.Update,
						fmt.Errorf("error_setting_user_state: %d, %w", from.Id, err))
					return
				}
				e.processStaticHandler(e.staticHandlers[nextState], client, &StateUpdate[User]{
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

func (e *engineWithPrivateStateHandlers[User]) switchState(
	nextState string, client *tgbotapi.TelegramBot, ctx context.Context, user User, update tgbotapi.Update) error {

	from := update.From()

	if handler := e.staticHandlers[nextState]; handler != nil {
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

func (e *engineWithPrivateStateHandlers[User]) processUserState(update tgbotapi.Update) (User, string, error) {
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
