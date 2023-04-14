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

	buttons []string

	buttonFormation []int
	maxButtonPerRow int
}

/*
buy:1234
-> open a link
-> send a new message
-> edit current message
-> reply with callback
*/

func NewInlineMenu[User any]() *InlineMenu[User] {
	return &InlineMenu[User]{
		buttonData:        make(map[string]string),
		buttonAlerts:      make(map[string]*buttonAlert),
		buttonUrls:        make(map[string]string),
		buttonInlineMenus: make(map[string]*buttonInlineMenu),
	}
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

// SetButtonFormation sets the button formation.
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

// getInlineKeyboardMarkup returns the inline keyboard markup.
func (i *InlineMenu[User]) getInlineKeyboardMarkup() *structs.InlineKeyboardMarkup {
	var rows [][]map[string]string
	var row []map[string]string

	for _, button := range i.buttons {
		if data, ok := i.buttonData[button]; ok {
			row = append(row, map[string]string{
				"text":          button,
				"callback_data": data,
			})
		} else if ba := i.getButtonAlertByButton(button); ba != nil {
			row = append(row, map[string]string{
				"text":          button,
				"callback_data": ba.data,
			})
		} else if url, ok := i.buttonUrls[button]; ok {
			row = append(row, map[string]string{
				"text": button,
				"url":  url,
			})
		} else if bim := i.getInlineMenuByButton(button); bim != nil {
			row = append(row, map[string]string{
				"text":          button,
				"callback_data": bim.data,
			})
		}

		cond := len(row) >= 2

		if i.maxButtonPerRow > 0 {
			cond = len(row) >= i.maxButtonPerRow
		}

		if len(i.buttonFormation) > len(rows) {
			cond = len(row) >= i.buttonFormation[len(rows)]
		}

		if cond {
			rows = append(rows, row)
			row = nil
		}
	}

	if len(row) > 0 {
		rows = append(rows, row)
	}

	return tools.Keyboards{}.NewInlineKeyboardFromSlicesOfMaps(rows)
}

// getButtonAlertByCommand returns the button alert by command.
func (i *InlineMenu[User]) getButtonAlertByButton(btn string) *buttonAlert {
	for key := range i.buttonAlerts {
		ba := i.buttonAlerts[key]

		if ba.button == btn {
			return ba
		}
	}

	return nil
}

// getInlineMenuByButton returns the inline menu by button.
func (i *InlineMenu[User]) getInlineMenuByButton(btn string) *buttonInlineMenu {
	for key := range i.buttonInlineMenus {
		ba := i.buttonInlineMenus[key]

		if ba.button == btn {
			return ba
		}
	}

	return nil
}
