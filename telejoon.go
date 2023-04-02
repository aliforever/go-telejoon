package telejoon

import (
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"strings"
)

type Telejoon[T any] struct {
	opts []*Options

	updates chan tgbotapi.Update

	handlers *Handlers[T]
}

func New[T any](updates chan tgbotapi.Update, handlers *Handlers[T], opts ...*Options) *Telejoon[T] {
	return &Telejoon[T]{
		opts:     opts,
		updates:  updates,
		handlers: handlers,
	}
}

func (t *Telejoon[T]) Start() {
	for update := range t.updates {
		if t.handlers != nil {
			if t.handlers.stateHandlers != nil && update.Message != nil && update.Message.IsPrivate() {
				go t.processPrivateMessage(update)
				continue
			}

			if t.handlers.callbackHandlers != nil && update.CallbackQuery != nil {
				go t.processCallback(update)
				continue
			}
		}
	}
}

func (t *Telejoon[T]) onErr(update tgbotapi.Update, err error) {
	if len(t.opts) > 0 && t.opts[0].onErr != nil {
		t.opts[0].onErr(update, err)
	}
}

func (t *Telejoon[T]) processCallback(update tgbotapi.Update) {
	command, args := splitCallbackData(update.CallbackQuery.Data, t.handlers.callbackHandlers.commandSeparator)

	handler := t.handlers.callbackHandlers.GetHandler(command)
	if handler == nil {
		t.onErr(update, fmt.Errorf("empty_handler_for_callback: %s", update.CallbackQuery.Data))
		return
	}

	user, err := t.handlers.stateHandlers.userRepository.Find(update.CallbackQuery.From.Id)
	if err != nil {
		user, err = t.handlers.stateHandlers.userRepository.Store(update.CallbackQuery.From)
		if err != nil {
			t.onErr(update, fmt.Errorf("store_user: %s", err))
			return
		}
	}

	handler(user, update, args...)
}

func splitCallbackData(data, separator string) (command string, args []string) {
	split := strings.Split(data, separator)
	if len(split) == 0 {
		return "", nil
	}

	return split[0], split[1:]
}

func (t *Telejoon[T]) processPrivateMessage(update tgbotapi.Update) {
	if t.handlers.stateHandlers.defaultState == "" {
		t.onErr(update, fmt.Errorf("empty_user_state"))
		return
	}

	userState, err := t.handlers.stateHandlers.userStateRepository.Find(update.Message.From.Id)
	if err != nil {
		err = t.handlers.stateHandlers.userStateRepository.Store(update.Message.From.Id, t.handlers.stateHandlers.defaultState)
		if err != nil {
			t.onErr(update, fmt.Errorf("store_user_state: %w", err))
			return
		}
	}

	if userState == "" {
		if t.handlers.stateHandlers.defaultState == "" {
			t.onErr(update, fmt.Errorf("empty_user_state"))
			return
		}
		err = t.handlers.stateHandlers.userStateRepository.Store(
			update.Message.From.Id, t.handlers.stateHandlers.defaultState)
		if err != nil {
			t.onErr(update, fmt.Errorf("store_user_state: %w", err))
			return
		}
	}

	user, err := t.handlers.stateHandlers.userRepository.Find(update.Message.From.Id)
	if err != nil {
		user, err = t.handlers.stateHandlers.userRepository.Store(update.Message.From)
		if err != nil {
			t.onErr(update, fmt.Errorf("store_user: %w", err))
			return
		}
	}

	handler := t.handlers.stateHandlers.GetHandler(userState)
	if handler == nil {
		t.onErr(update, fmt.Errorf("empty_handler_for_state: %s", userState))
		return
	}

	if nextState := handler(user, update, false); nextState != "" {
		err = t.handlers.stateHandlers.userStateRepository.Store(update.Message.From.Id, nextState)
		if err != nil {
			t.onErr(update, fmt.Errorf("store_user_state: %s", err))
			return
		}

		handler = t.handlers.stateHandlers.GetHandler(nextState)
		if handler == nil {
			t.onErr(update, fmt.Errorf("empty_handler_for_state: %s", nextState))
			return
		}

		_ = handler(user, update, true)
	}
}
