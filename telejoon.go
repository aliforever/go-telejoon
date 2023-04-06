package telejoon

import (
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"strings"
)

type Telejoon[User UserI, Language LanguageI] struct {
	opts []*Options

	client *tgbotapi.TelegramBot

	handlers *Handlers[User, Language]

	languages map[string]Language
}

func New[User UserI, Language LanguageI](
	client *tgbotapi.TelegramBot, handlers *Handlers[User, Language], languages []Language,
	opts ...*Options) *Telejoon[User, Language] {

	return &Telejoon[User, Language]{
		opts:     opts,
		client:   client,
		handlers: handlers,
		languages: func() map[string]Language {
			langs := make(map[string]Language)
			for _, lang := range languages {
				langs[lang.Code()] = lang
			}
			return langs
		}(),
	}
}

func (t *Telejoon[User, Language]) Start() {
	for update := range t.client.Updates() {
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

func (t *Telejoon[User, Language]) DefaultWelcomeState(update StateUpdate[User, Language]) string {
	if update.User.LanguageCode() == "" {
		return "DefaultChangeLanguageState"
	}

	userLanguage, err := t.getUserLanguage(update)
	if err != nil {
		t.onErr(update.Update, err)
		return ""
	}

	if !update.IsSwitched {
		if update.Update.Message.Text == userLanguage.SelectLanguage() {
			return "DefaultChangeLanguageState"
		}
	}

	keyboard := t.client.Tools.Keyboards.NewReplyKeyboardFromSlicesOfStrings([][]string{
		{userLanguage.SelectLanguage()},
	})

	t.client.Send(t.client.Message().SetChatId(update.Update.From().Id).
		SetText(userLanguage.Welcome()).
		SetReplyMarkup(keyboard))

	return ""
}

func (t *Telejoon[User, Language]) DefaultChangeLanguageState(update StateUpdate[User, Language]) string {
	var keyboard *structs.ReplyKeyboardMarkup

	{
		rows := [][]string{}
		row := []string{}

		for _, language := range t.languages {
			row = append(row, fmt.Sprintf("%s %s", language.Flag(), language.Name()))
			if len(row) == 2 {
				rows = append(rows, row)
				row = []string{}
			}
		}

		if len(row) > 0 {
			rows = append(rows, row)
		}

		keyboard = t.client.Tools.Keyboards.NewReplyKeyboardFromSlicesOfStrings(rows)
	}

	if !update.IsSwitched {
		for _, language := range t.languages {
			if update.Update.Message.Text == fmt.Sprintf("%s %s", language.Flag(), language.Name()) {
				err := t.handlers.stateHandlers.userRepository.SetLanguage(update.Update.From().Id, language.Code())
				if err != nil {
					t.onErr(update.Update, err)
					return ""
				}

				return "DefaultWelcomeState"
			}
		}
	}

	var text string

	{
		texts := []string{}

		for _, language := range t.languages {
			texts = append(texts, fmt.Sprintf("%s %s", language.Flag(), language.SelectLanguage()))
		}

		text = strings.Join(texts, "\n")
	}

	_, err := t.client.Send(t.client.Message().
		SetChatId(update.Update.From().Id).
		SetText(text).
		SetReplyMarkup(keyboard))
	if err != nil {
		t.onErr(update.Update, err)
	}

	return ""
}

func (t *Telejoon[User, Language]) getUserLanguage(update StateUpdate[User, Language]) (Language, error) {
	if len(t.languages) == 0 {
		return nil, fmt.Errorf("empty_languages")
	}

	if lang, ok := t.languages[update.User.LanguageCode()]; ok {
		return lang, nil
	}

	return nil, fmt.Errorf("language_not_found: %s", update.User.LanguageCode())
}

func (t *Telejoon[User, Language]) onErr(update tgbotapi.Update, err error) {
	if len(t.opts) > 0 && t.opts[0].onErr != nil {
		t.opts[0].onErr(update, err)
	}
}

func (t *Telejoon[User, Language]) processCallback(update tgbotapi.Update) {
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

	if user.LanguageCode() == "" {
		t.onErr(update, fmt.Errorf("empty_language_code_for_user: %+v", user))
		return
	}

	lang, ok := t.languages[user.LanguageCode()]
	if !ok {
		t.onErr(update, fmt.Errorf("language_not_found: %s", user.LanguageCode()))
		return
	}

	handler(CallbackUpdate[User, Language]{
		User:     user,
		Language: lang,
		Update:   update,
	}, args...)
}

func splitCallbackData(data, separator string) (command string, args []string) {
	split := strings.Split(data, separator)
	if len(split) == 0 {
		return "", nil
	}

	return split[0], split[1:]
}

func (t *Telejoon[User, Language]) processPrivateMessage(update tgbotapi.Update) {
	// check if default state is set
	if t.handlers.stateHandlers.defaultState == "" {
		t.onErr(update, fmt.Errorf("empty_user_state"))
		return
	}

	userState, err := t.handlers.stateHandlers.userStateRepository.Find(update.Message.From.Id)
	if err != nil || userState == "" {
		userState = t.handlers.stateHandlers.defaultState
		err = t.handlers.stateHandlers.userStateRepository.Store(update.Message.From.Id, userState)
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

	if user.LanguageCode() == "" {
		t.onErr(update, fmt.Errorf("empty_language_code_for_user: %+v", user))
		return
	}

	lang, ok := t.languages[user.LanguageCode()]
	if !ok {
		t.onErr(update, fmt.Errorf("language_not_found: %s", user.LanguageCode()))
		return
	}

	handler := t.handlers.stateHandlers.GetHandler(userState)
	if handler == nil {
		t.onErr(update, fmt.Errorf("empty_handler_for_state: %s", userState))
		return
	}

	if nextState := handler(StateUpdate[User, Language]{
		User:       user,
		Language:   lang,
		Update:     update,
		IsSwitched: false,
	}); nextState != "" {
		err = t.handlers.stateHandlers.userStateRepository.Update(update.Message.From.Id, nextState)
		if err != nil {
			t.onErr(update, fmt.Errorf("update_user_state: %s", err))
			return
		}

		handler = t.handlers.stateHandlers.GetHandler(nextState)
		if handler == nil {
			t.onErr(update, fmt.Errorf("empty_handler_for_state: %s", nextState))
			return
		}

		_ = handler(StateUpdate[User, Language]{
			User:       user,
			Language:   lang,
			Update:     update,
			IsSwitched: true,
		})
	}
}
