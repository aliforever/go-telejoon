package telejoon

import (
	"fmt"
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

type baseInlineButton struct {
	button string
	data   string

	options []*ButtonOptions
}

// Button returns the button
func (t baseInlineButton) Button() string {
	return t.button
}

// Data returns the data
func (t baseInlineButton) Data() string {
	return t.data
}

// Options returns the options
func (t baseInlineButton) Options() *ButtonOptions {
	if len(t.options) == 0 {
		return nil
	}

	return t.options[0]
}

type inlineUrlButton struct {
	baseInlineButton
}

type inlineInlineMenuButton struct {
	baseInlineButton

	menu string
	edit bool
}

type inlineAlertButton struct {
	baseInlineButton

	text      string
	showAlert bool
}

type inlineStateButton struct {
	baseInlineButton

	state string
}

type inlineCallbackButton struct {
	baseInlineButton
}

type InlineAction interface {
	Button() string
	Data() string
	Options() *ButtonOptions
}

type InlineActionBuilder struct {
	locker sync.Mutex

	inlineMenu string

	buttons []InlineAction

	buttonFormation []int
	maxButtonPerRow int
}

// NewInlineActionBuilder creates a new InlineActionBuilder.
func NewInlineActionBuilder() *InlineActionBuilder {
	return &InlineActionBuilder{}
}

// SetMaxButtonPerRow sets the maximum number of buttons per row.
func (b *InlineActionBuilder) SetMaxButtonPerRow(max int) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.maxButtonPerRow = max

	return b
}

func (b *InlineActionBuilder) SetButtonFormation(formation ...int) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttonFormation = formation

	return b
}

// AddUrlButton adds a new url button to the InlineActionBuilder.
func (b *InlineActionBuilder) AddUrlButton(
	button string, address string, opts ...*ButtonOptions) *InlineActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineUrlButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			data:    address,
			options: opts,
		},
	})

	return b
}

func (b *InlineActionBuilder) AddUrlButtonT(
	button string, address string, opts ...*ButtonOptions) *InlineActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineUrlButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			data:    address,
			options: defaultOptions,
		},
	})

	return b
}

func (b *InlineActionBuilder) AddInlineMenuButton(
	button, data, inlineMenu string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			data:    data,
			options: opts,
		},
		menu: inlineMenu,
	})

	return b
}

func (b *InlineActionBuilder) AddInlineMenuButtonWithEdit(
	button, data, inlineMenu string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: opts,
			data:    data,
		},
		menu: inlineMenu,
		edit: true,
	})

	return b
}

func (b *InlineActionBuilder) AddInlineMenuButtonT(
	button, data, inlineMenu string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			data:    data,
			options: defaultOptions,
		},
		menu: inlineMenu,
	})

	return b
}

func (b *InlineActionBuilder) AddInlineMenuButtonWithEditT(
	button, data, inlineMenu string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineInlineMenuButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: defaultOptions,
			data:    data,
		},
		menu: inlineMenu,
		edit: true,
	})

	return b
}

func (b *InlineActionBuilder) AddAlertButton(
	button, data, alertText string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineAlertButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: opts,
			data:    data,
		},
		text: alertText,
	})

	return b
}

func (b *InlineActionBuilder) AddAlertButtonWithDialog(
	button, data, alertText string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineAlertButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: defaultOptions,
			data:    data,
		},
		text:      alertText,
		showAlert: true,
	})

	return b
}

func (b *InlineActionBuilder) AddAlertButtonT(
	button, data, alertText string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineAlertButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: defaultOptions,
			data:    data,
		},
		text: alertText,
	})

	return b
}

func (b *InlineActionBuilder) AddAlertButtonWithDialogT(
	button, data, alertText string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineAlertButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: defaultOptions,
			data:    data,
		},
		text:      alertText,
		showAlert: true,
	})

	return b
}

func (b *InlineActionBuilder) AddStateButton(button, data, state string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineStateButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: opts,
			data:    data,
		},
		state: state,
	})

	return b
}

// AddStateButtonT adds a state action to the InlineActionBuilder with name translation.
func (b *InlineActionBuilder) AddStateButtonT(button, data, state string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineStateButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			data:    data,
			options: defaultOptions,
		},
		state: state,
	})

	return b
}

func (b *InlineActionBuilder) AddCallbackButton(button, data string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineCallbackButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: opts,
			data:    data,
		},
	})

	return b
}

// AddCallbackButtonT adds a state action to the InlineActionBuilder with name translation.
func (b *InlineActionBuilder) AddCallbackButtonT(button, data string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		defaultOptions = append(defaultOptions, opts...)
	}

	b.buttons = append(b.buttons, inlineCallbackButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			data:    data,
			options: defaultOptions,
		},
	})

	return b
}

// buildButtons builds the buttons.
func (b *InlineActionBuilder) buildButtons(language *Language) *structs.InlineKeyboardMarkup {
	if len(b.buttons) == 0 {
		return nil
	}

	var rows []map[string]string

	for _, button := range b.buttons {
		name := button.Button()

		shouldBreakAfter := false

		if opts := button.Options(); opts != nil {
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

		if val, ok := button.(inlineUrlButton); ok {
			row["url"] = val.data
		} else {
			row["callback_data"] = fmt.Sprintf("%s:%s", b.inlineMenu, button.Data())
		}

		rows = append(rows, row)

		if shouldBreakAfter {
			rows = append(rows, nil)
		}
	}

	return tools.Keyboards{}.NewInlineKeyboardFromSlicesOfMapWithFormation(rows, b.maxButtonPerRow, b.buttonFormation)
}

func (b *InlineActionBuilder) getByCallbackActionData() map[string]InlineAction {
	b.locker.Lock()
	defer b.locker.Unlock()

	var data = make(map[string]InlineAction)

	for _, button := range b.buttons {
		if _, ok := button.(inlineUrlButton); !ok {
			sp := strings.Split(button.Data(), ":")
			data[sp[0]] = button
		}
	}

	return data
}
