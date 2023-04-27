package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
	"sync"
)

type buttonAlert struct {
	data      string
	button    string
	text      string
	showAlert bool
}

type buttonInlineMenu struct {
	data   string
	button string
	edit   bool
	args   []string
}

type InlineMenu[User any] struct {
	lock sync.Mutex

	replyText string

	replyWithFunc func(*tgbotapi.TelegramBot, *CallbackUpdate[User])

	buttonData map[string]string
	buttonUrls map[string]string

	buttonAlerts      map[string]*buttonAlert
	buttonInlineMenus map[string]*buttonInlineMenu

	buttons            []string
	languageKeyButtons map[string]bool

	buttonFormation []int
	maxButtonPerRow int

	middlewares []func(*tgbotapi.TelegramBot, *CallbackUpdate[User]) bool
}

func NewInlineMenu[User any]() *InlineMenu[User] {
	return &InlineMenu[User]{
		buttonData:         make(map[string]string),
		buttonAlerts:       make(map[string]*buttonAlert),
		buttonUrls:         make(map[string]string),
		buttonInlineMenus:  make(map[string]*buttonInlineMenu),
		languageKeyButtons: make(map[string]bool),
	}
}

// AddMiddleware adds a middleware to the inline menu
func (i *InlineMenu[User]) AddMiddleware(middleware func(*tgbotapi.TelegramBot, *CallbackUpdate[User]) bool) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.middlewares = append(i.middlewares, middleware)

	return i
}

func (i *InlineMenu[User]) AddButtonData(button, data string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonData[button] = data

	i.buttons = append(i.buttons, button)

	return i
}

func (i *InlineMenu[User]) AddButtonUrl(button, url string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonUrls[button] = url

	i.buttons = append(i.buttons, button)

	return i
}

func (i *InlineMenu[User]) AddButtonInlineMenu(button, menu string, edit bool, args ...string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonInlineMenus[menu] = &buttonInlineMenu{
		button: button,
		data:   menu,
		edit:   edit,
		args:   args,
	}

	i.buttons = append(i.buttons, button)

	return i
}

func (i *InlineMenu[User]) AddDataButtonAlert(button, data string, text string, showAlert bool) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonAlerts[data] = &buttonAlert{
		data:      data,
		text:      text,
		button:    button,
		showAlert: showAlert,
	}

	i.buttons = append(i.buttons, button)

	return i
}

func (i *InlineMenu[User]) AddLanguageKeyButtonData(button, data string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonData[button] = data

	i.buttons = append(i.buttons, button)

	i.languageKeyButtons[button] = true

	return i
}

func (i *InlineMenu[User]) AddLanguageKeyButtonUrl(button, url string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonUrls[button] = url

	i.buttons = append(i.buttons, button)

	i.languageKeyButtons[button] = true

	return i
}

func (i *InlineMenu[User]) AddLanguageKeyButtonInlineMenu(
	button, menu string, edit bool, args ...string) *InlineMenu[User] {

	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonInlineMenus[menu] = &buttonInlineMenu{
		button: button,
		data:   menu,
		edit:   edit,
		args:   args,
	}

	i.buttons = append(i.buttons, button)

	i.languageKeyButtons[button] = true

	return i
}

func (i *InlineMenu[User]) AddLanguageKeyDataButtonAlert(
	button, data string, text string, showAlert bool) *InlineMenu[User] {

	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonAlerts[data] = &buttonAlert{
		data:      data,
		text:      text,
		button:    button,
		showAlert: showAlert,
	}

	i.buttons = append(i.buttons, button)

	i.languageKeyButtons[button] = true

	return i
}

func (i *InlineMenu[User]) AddReplyText(text string) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.replyText = text

	return i
}

func (i *InlineMenu[User]) AddReplyWithFunc(
	f func(*tgbotapi.TelegramBot, *CallbackUpdate[User])) *InlineMenu[User] {

	i.lock.Lock()
	defer i.lock.Unlock()

	i.replyWithFunc = f

	return i
}

// SetButtonFormation sets the action formation.
// The formation is a slice of int, each representing number of buttons in a row.
func (i *InlineMenu[User]) SetButtonFormation(formation ...int) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.buttonFormation = formation

	return i
}

// SetMaxButtonPerRow sets the maximum number of buttons per row.
func (i *InlineMenu[User]) SetMaxButtonPerRow(max int) *InlineMenu[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.maxButtonPerRow = max

	return i
}

// GetInlineKeyboardMarkup returns the inline keyboard markup.
func (i *InlineMenu[User]) GetInlineKeyboardMarkup(language *Language) *structs.InlineKeyboardMarkup {
	var row []map[string]string

	for _, button := range i.buttons {
		buttonText := button

		if language != nil {
			if i.languageKeyButtons[button] {
				translatedText, err := language.Get(button)
				if err == nil {
					buttonText = translatedText
				}
			}
		}

		if data, ok := i.buttonData[button]; ok {
			row = append(row, map[string]string{
				"textHandler":   buttonText,
				"callback_data": data,
			})
		} else if ba := i.getButtonAlertByButton(button); ba != nil {
			row = append(row, map[string]string{
				"textHandler":   buttonText,
				"callback_data": ba.data,
			})
		} else if url, ok := i.buttonUrls[button]; ok {
			row = append(row, map[string]string{
				"textHandler": buttonText,
				"url":         url,
			})
		} else if bim := i.getInlineMenuByButton(button); bim != nil {
			row = append(row, map[string]string{
				"textHandler":   buttonText,
				"callback_data": bim.data,
			})
		}
	}

	return tools.Keyboards{}.NewInlineKeyboardFromSlicesOfMapWithFormation(
		row, i.maxButtonPerRow, i.buttonFormation)
}

// GetReplyText returns the reply text.
func (i *InlineMenu[User]) GetReplyText(language *Language) string {
	if language != nil {
		translatedText, err := language.Get(i.replyText)
		if err == nil {
			return translatedText
		}
	}

	return i.replyText
}

// GetReplyWithFunc returns the reply with func.
func (i *InlineMenu[User]) GetReplyWithFunc() func(*tgbotapi.TelegramBot, *CallbackUpdate[User]) {
	return i.replyWithFunc
}

// getButtonAlertByCommand returns the action alert by command.
func (i *InlineMenu[User]) getButtonAlertByButton(btn string) *buttonAlert {
	for key := range i.buttonAlerts {
		ba := i.buttonAlerts[key]

		if ba.button == btn {
			return ba
		}
	}

	return nil
}

// getInlineMenuByButton returns the inline menu by action.
func (i *InlineMenu[User]) getInlineMenuByButton(btn string) *buttonInlineMenu {
	for key := range i.buttonInlineMenus {
		ba := i.buttonInlineMenus[key]

		if ba.button == btn {
			return ba
		}
	}

	return nil
}

// getMiddlewares returns the middlewares.
func (i *InlineMenu[User]) getMiddlewares() []func(bot *tgbotapi.TelegramBot, update *CallbackUpdate[User]) bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.middlewares
}
