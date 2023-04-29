package telejoon

import (
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
	"strings"
	"sync"
)

type InlineActionKind string

const (
	InlineActionKindUrl        InlineActionKind = "URL"
	InlineActionKindInlineMenu InlineActionKind = "INLINE_MENU"
	InlineActionKindAlert      InlineActionKind = "ALERT"
	InlineActionKindState      InlineActionKind = "STATE"
	InlineActionKindCallback   InlineActionKind = "CALLBACK"
)

type inlineUrlButton struct {
	baseButton

	address string
}

func (t inlineUrlButton) Name() string {
	return t.button
}

func (t inlineUrlButton) Kind() InlineActionKind {
	return InlineActionKindUrl
}

func (t inlineUrlButton) Result() string {
	return t.address
}

type inlineInlineMenuButton struct {
	baseButton

	menuName string
}

func (t inlineInlineMenuButton) Name() string {
	return t.button
}

func (t inlineInlineMenuButton) Kind() InlineActionKind {
	return InlineActionKindInlineMenu
}

func (t inlineInlineMenuButton) Result() string {
	return t.menuName
}

type inlineAlertButton struct {
	baseButton

	text      string
	showAlert bool
}

func (t inlineAlertButton) Name() string {
	return t.button
}

func (t inlineAlertButton) Kind() InlineActionKind {
	return InlineActionKindAlert
}

func (t inlineAlertButton) Result() string {
	return t.text
}

type inlineStateButton struct {
	baseButton

	state string
}

func (t inlineStateButton) Name() string {
	return t.button
}

func (t inlineStateButton) Kind() InlineActionKind {
	return InlineActionKindState
}

func (t inlineStateButton) Result() string {
	return t.state
}

type inlineCallbackButton struct {
	baseButton

	data string
}

func (t inlineCallbackButton) Name() string {
	return t.button
}

func (t inlineCallbackButton) Kind() InlineActionKind {
	return InlineActionKindCallback
}

func (t inlineCallbackButton) Result() string {
	return t.data
}

type InlineAction interface {
	Name() string
	Kind() InlineActionKind
	Result() string
}

type inlineActionBuilder struct {
	locker sync.Mutex

	buttons []InlineAction

	buttonOptions map[string][]*ButtonOptions

	buttonFormation []int
	maxButtonPerRow int
}

// NewInlineActionBuilder creates a new inlineActionBuilder.
func NewInlineActionBuilder() *inlineActionBuilder {
	return &inlineActionBuilder{
		buttonOptions: make(map[string][]*ButtonOptions),
	}
}

// SetMaxButtonPerRow sets the maximum number of buttons per row.
func (b *inlineActionBuilder) SetMaxButtonPerRow(max int) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.maxButtonPerRow = max

	return b
}

func (b *inlineActionBuilder) SetButtonFormation(formation ...int) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttonFormation = formation

	return b
}

// AddUrlButton adds a new url button to the inlineActionBuilder.
func (b *inlineActionBuilder) AddUrlButton(
	button string, address string, opts ...*ButtonOptions) *inlineActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineUrlButton{
		baseButton: baseButton{
			button: button,
		},
		address: address,
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	return b
}

func (b *inlineActionBuilder) AddUrlButtonT(
	button string, address string, opts ...*ButtonOptions) *inlineActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineUrlButton{
		baseButton: baseButton{
			button: button,
		},
		address: address,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddInlineMenuButton(
	button, inlineMenu string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		menuName: inlineMenu,
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	return b
}

func (b *inlineActionBuilder) AddInlineMenuButtonWithEdit(
	button, inlineMenu string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		menuName: inlineMenu,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().ShouldEdit(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddInlineMenuButtonT(
	button, inlineMenu string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		menuName: inlineMenu,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddInlineMenuButtonWithEditT(
	button, inlineMenu string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		menuName: inlineMenu,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName().ShouldEdit(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddAlertButton(
	button, callbackData, alertText string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineAlertButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: callbackData,
	})

	btnOptions := NewButtonOptions().Alert(alertText)

	b.buttonOptions[button] = []*ButtonOptions{
		btnOptions,
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddAlertButtonWithDialog(
	button, callbackData, alertText string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineAlertButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: callbackData,
	})

	btnOptions := NewButtonOptions().Alert(alertText).ShowAlertDialog()

	b.buttonOptions[button] = []*ButtonOptions{
		btnOptions,
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddAlertButtonT(
	button, callbackData, alertText string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineAlertButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: callbackData,
	})

	btnOptions := NewButtonOptions().TranslateName().Alert(alertText)

	b.buttonOptions[button] = []*ButtonOptions{
		btnOptions,
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddAlertButtonWithDialogT(
	button, callbackData, alertText string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineAlertButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: callbackData,
	})

	btnOptions := NewButtonOptions().TranslateName().Alert(alertText).ShowAlertDialog()

	b.buttonOptions[button] = []*ButtonOptions{
		btnOptions,
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddStateButton(button, state string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineStateButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		}, state: state,
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// AddStateButtonT adds a state action to the inlineActionBuilder with name translation.
func (b *inlineActionBuilder) AddStateButtonT(button, state string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineStateButton{
		baseButton: baseButton{
			button: button,
		}, state: state,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

func (b *inlineActionBuilder) AddCallbackButton(button, data string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineCallbackButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		}, data: data,
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// AddCallbackButtonT adds a state action to the inlineActionBuilder with name translation.
func (b *inlineActionBuilder) AddCallbackButtonT(button, data string, opts ...*ButtonOptions) *inlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineCallbackButton{
		baseButton: baseButton{
			button: button,
		}, data: data,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// buildButtons builds the buttons.
func (b *inlineActionBuilder) buildButtons(language *Language) *structs.InlineKeyboardMarkup {
	if len(b.buttons) == 0 {
		return nil
	}

	var rows []map[string]string

	for _, button := range b.buttons {
		name := button.Name()

		shouldBreakAfter := false

		if opts := b.getOptionsForButton(name); opts != nil {
			if language != nil && opts.translateName {
				btnTxt, _ := language.Get(name)
				if btnTxt != "" {
					name = btnTxt
				}
			}

			if opts.breakBefore {
				rows = append(rows, nil)
			}

			if opts.breakAfter {
				shouldBreakAfter = true
			}
		}

		row := map[string]string{
			"text": name,
		}

		if button.Kind() == InlineActionKindUrl {
			row["url"] = button.Result()
		} else {
			row["callback_data"] = button.Result()
		}

		rows = append(rows, row)

		if shouldBreakAfter {
			rows = append(rows, nil)
		}
	}

	return tools.Keyboards{}.NewInlineKeyboardFromSlicesOfMapWithFormation(rows, b.maxButtonPerRow, b.buttonFormation)
}

func (b *inlineActionBuilder) getOptionsForButton(button string) *ButtonOptions {
	b.locker.Lock()
	defer b.locker.Unlock()

	if opts := b.buttonOptions[button]; len(opts) == 0 {
		return nil
	}

	return b.buttonOptions[button][0]
}

func (b *inlineActionBuilder) getByCallbackActionData() map[string]InlineAction {
	b.locker.Lock()
	defer b.locker.Unlock()

	var data = make(map[string]InlineAction)

	for _, button := range b.buttons {
		if button.Kind() != InlineActionKindUrl {
			sp := strings.Split(button.Result(), ":")
			data[sp[0]] = button
		}
	}

	return data
}
