package telejoon

import (
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
	"strings"
	"sync"
)

type baseInlineButton struct {
	button TextBuilder
	data   string

	options []*ButtonOptions
}

func (t baseInlineButton) Name(update *StateUpdate) string {
	return t.button.String(update)
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

	handler CallbackHandler
}

type InlineAction interface {
	Name(update *StateUpdate) string
	Data() string
	Options() *ButtonOptions
}

type InlineActionBuilderKind interface {
	Build(update *StateUpdate) *InlineActionBuilder
}

type DeferredInlineActionBuilder func(update *StateUpdate) *InlineActionBuilder

// Build builds the deferred action builder.
func (d DeferredInlineActionBuilder) Build(update *StateUpdate) *InlineActionBuilder {
	return d(update)
}

// NewDeferredInlineActionBuilder creates a new DeferredInlineActionBuilder.
func NewDeferredInlineActionBuilder(builder func(update *StateUpdate) *InlineActionBuilder) DeferredInlineActionBuilder {
	return DeferredInlineActionBuilder(builder)
}

type CallbackHandler func(
	client *tgbotapi.TelegramBot,
	update *StateUpdate,
	args ...string,
) (SwitchAction, error)

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
	button TextBuilder, address string, opts ...*ButtonOptions) *InlineActionBuilder {

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

func (b *InlineActionBuilder) AddInlineMenuButton(
	button TextBuilder, data, inlineMenu string, opts ...*ButtonOptions) *InlineActionBuilder {
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
	button TextBuilder, data, inlineMenu string, opts ...*ButtonOptions) *InlineActionBuilder {
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

func (b *InlineActionBuilder) AddAlertButton(
	button TextBuilder, data, alertText string, opts ...*ButtonOptions) *InlineActionBuilder {
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
	button TextBuilder, data, alertText string, opts ...*ButtonOptions) *InlineActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	defaultOptions := []*ButtonOptions{}

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

func (b *InlineActionBuilder) AddStateButton(button TextBuilder, data, state string, opts ...*ButtonOptions) *InlineActionBuilder {
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

func (b *InlineActionBuilder) AddCallbackButton(
	button TextBuilder,
	data string,
	handler CallbackHandler,
	opts ...*ButtonOptions,
) *InlineActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineCallbackButton{
		baseInlineButton: baseInlineButton{
			button:  button,
			options: opts,
			data:    data,
		},
		handler: handler,
	})

	return b
}

func (b *InlineActionBuilder) Build(_ *StateUpdate) *InlineActionBuilder {
	return b
}

// buildButtons builds the buttons.
func (b *InlineActionBuilder) buildButtons(update *StateUpdate) *structs.InlineKeyboardMarkup {
	if len(b.buttons) == 0 {
		return nil
	}

	var rows []map[string]string

	for _, button := range b.buttons {
		name := button.Name(update)

		shouldBreakAfter := false

		if opts := button.Options(); opts != nil {
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
