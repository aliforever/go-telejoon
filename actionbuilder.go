package telejoon

import (
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
	"sync"
)

type (
	ActionKind string
)

const (
	ActionKindText       ActionKind = "TEXT"
	ActionKindInlineMenu ActionKind = "INLINE_MENU"
	ActionKindState      ActionKind = "STATE"
	ActionKindRaw        ActionKind = "RAW"
)

type Action interface {
	Name(update *StateUpdate) string
}

type baseCommand struct {
	command TextBuilder
}

func (b baseCommand) Name(update *StateUpdate) string {
	return b.command.String(update)
}

// textCommand is a command that sends a text message.
type textCommand struct {
	baseCommand

	text string
}

// inlineMenuCommand is a command that is used to switch to an inline menu.
type inlineMenuCommand struct {
	baseCommand

	inlineMenu string
}

// stateCommand is a command that is used to switch to a state.
type stateCommand struct {
	baseCommand

	state string
}

type baseButtonOptions interface {
	Options() *ButtonOptions
	CanBeShown(update *StateUpdate) bool
}

type baseButton struct {
	button TextBuilder

	condition func(update *StateUpdate) bool

	options []*ButtonOptions
}

func (t baseButton) Name(update *StateUpdate) string {
	return t.button.String(update)
}

func (t baseButton) Options() *ButtonOptions {
	if len(t.options) == 0 {
		return nil
	}

	return t.options[0]
}

func (t baseButton) CanBeShown(update *StateUpdate) bool {
	return t.condition == nil || t.condition(update)
}

// textButton is a button that sends a text message when clicked.
type textButton struct {
	baseButton

	text TextBuilder
}

// inlineMenuButton is a button that switches to an inline menu when clicked.
type inlineMenuButton struct {
	baseButton

	inlineMenu string
}

// stateButton is a button that switches to a state when clicked.
type stateButton struct {
	baseButton

	state string
	hook  UpdateHandler
}

// rawButton is only a raw button that does nothing but sends the button name.
type rawButton struct {
	baseButton
}

type ActionBuilderKind interface {
	build(update *StateUpdate) *ActionBuilder
}

type DeferredActionBuilder func(update *StateUpdate) *ActionBuilder

// Build builds the deferred action builder.
func (d DeferredActionBuilder) build(update *StateUpdate) *ActionBuilder {
	return d(update)
}

// NewDeferredActionBuilder creates a new DeferredActionBuilder.
func NewDeferredActionBuilder(
	builder func(update *StateUpdate) *ActionBuilder,
) DeferredActionBuilder {

	return builder
}

type ActionBuilder struct {
	locker sync.Mutex

	buttons  []Action
	commands []Action

	buttonFormation []int
	maxButtonPerRow int
}

// NewStaticActionBuilder creates a new ActionBuilder.
func NewStaticActionBuilder() *ActionBuilder {
	return &ActionBuilder{}
}

// SetMaxButtonPerRow sets the maximum number of buttons per row.
func (b *ActionBuilder) SetMaxButtonPerRow(max int) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.maxButtonPerRow = max

	return b
}

func (b *ActionBuilder) SetButtonFormation(formation ...int) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttonFormation = formation

	return b
}

// AddTextButton adds a textHandler action to the ActionBuilder.
func (b *ActionBuilder) AddTextButton(button TextBuilder, text TextBuilder, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: text,
	})

	return b
}

// AddConditionalTextButton adds a textHandler action to the ActionBuilder with a condition.
func (b *ActionBuilder) AddConditionalTextButton(
	cond func(update *StateUpdate) bool,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:    button,
			condition: cond,
			options:   opts,
		},
		text: text,
	})

	return b
}

// AddInlineMenuButton adds an inline menu action to the ActionBuilder.
func (b *ActionBuilder) AddInlineMenuButton(
	button TextBuilder,
	inlineMenu string,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		inlineMenu: inlineMenu,
	})

	return b
}

// AddStateButton adds a state action to the ActionBuilder.
func (b *ActionBuilder) AddStateButton(button TextBuilder, state string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, stateButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		}, state: state,
	})

	return b
}

// AddStateButtonWithHook adds a state action to the ActionBuilder with hook.
func (b *ActionBuilder) AddStateButtonWithHook(
	button TextBuilder,
	state string,
	hook UpdateHandler,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, stateButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		state: state,
		hook:  hook,
	})

	return b
}

// AddRawButton adds a raw button to the ActionBuilder.
func (b *ActionBuilder) AddRawButton(button TextBuilder, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, rawButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
	})

	return b
}

// AddTextCommand adds a textHandler command to the ActionBuilder.
func (b *ActionBuilder) AddTextCommand(command TextBuilder, text string) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.commands = append(b.commands, textCommand{
		baseCommand: baseCommand{
			command: command,
		},
		text: text,
	})

	return b
}

// AddInlineMenuCommand adds an inline menu command to the ActionBuilder.
func (b *ActionBuilder) AddInlineMenuCommand(command TextBuilder, inlineMenu string) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.commands = append(b.commands, inlineMenuCommand{
		baseCommand: baseCommand{
			command: command,
		},
		inlineMenu: inlineMenu,
	})

	return b
}

// AddStateCommand adds a state command to the ActionBuilder.
func (b *ActionBuilder) AddStateCommand(command TextBuilder, state string) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.commands = append(b.commands, stateCommand{
		baseCommand: baseCommand{
			command: command,
		},
		state: state,
	})

	return b
}

// AddCustomButton adds a custom action of button type to the ActionBuilder.
func (b *ActionBuilder) AddCustomButton(action Action) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, action)

	return b
}

// AddCustomCommand adds a custom action of command type to the ActionBuilder.
func (b *ActionBuilder) AddCustomCommand(action Action) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.commands = append(b.commands, action)

	return b
}

func (b *ActionBuilder) build(_ *StateUpdate) *ActionBuilder {
	return b
}

// getButtonByButton returns the action by the button.
func (b *ActionBuilder) getButtonByButton(update *StateUpdate, button string) Action {
	for _, btn := range b.buttons {
		if btn.Name(update) == button {
			return btn
		}
	}

	return nil
}

// buildButtons builds the buttons.
func (b *ActionBuilder) buildButtons(update *StateUpdate, reverseButtonOrderInRows bool) *structs.ReplyKeyboardMarkup {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(b.buttons) == 0 {
		return nil
	}

	newButtons := []string{}

	for _, button := range b.buttons {
		name := button.Name(update)

		shouldBreakAfter := false

		if opts, ok := button.(baseButtonOptions); ok {
			if !opts.CanBeShown(update) {
				continue
			}

			if btnOpts := opts.Options(); btnOpts != nil {
				if btnOpts.breakBefore {
					newButtons = append(newButtons, "")
				}

				if btnOpts.breakAfter {
					shouldBreakAfter = true
				}
			}
		}

		newButtons = append(newButtons, name)

		if shouldBreakAfter {
			newButtons = append(newButtons, "")
		}
	}

	return tools.Keyboards{}.NewReplyKeyboardFromSliceOfStringsWithFormation(
		newButtons,
		b.maxButtonPerRow,
		b.buttonFormation,
		reverseButtonOrderInRows,
	)
}
