package telejoon

type (
	ActionKind string
)

const (
	ActionKindText       ActionKind = "TEXT"
	ActionKindInlineMenu ActionKind = "INLINE_MENU"
	ActionKindState      ActionKind = "STATE"
)

type baseCommand struct {
	command string
}

// textCommand is a command that sends a text message.
type textCommand struct {
	baseCommand
	text string
}

func (t textCommand) Kind() ActionKind {
	return ActionKindText
}

func (t textCommand) Result() string {
	return t.text
}

// --------------------------------------------

// inlineMenuCommand is a command that is used to switch to an inline menu.
type inlineMenuCommand struct {
	baseCommand
	inlineMenu string
}

func (t inlineMenuCommand) Kind() ActionKind {
	return ActionKindInlineMenu
}

func (t inlineMenuCommand) Result() string {
	return t.inlineMenu
}

// --------------------------------------------

// stateCommand is a command that is used to switch to a state.
type stateCommand struct {
	baseCommand
	state string
}

func (t stateCommand) Kind() ActionKind {
	return ActionKindState
}

func (t stateCommand) Result() string {
	return t.state
}

// --------------------------------------------

type baseButton struct {
	button  string
	options []ButtonOptions
}

// textButton is a action that sends a text message.
type textButton struct {
	baseButton
	text string
}

func (t textButton) Kind() ActionKind {
	return ActionKindText
}

func (t textButton) Result() string {
	return t.text
}

// --------------------------------------------

// inlineMenuButton is a action that is used to switch to an inline menu.
type inlineMenuButton struct {
	baseButton
	inlineMenu string
}

func (t inlineMenuButton) Kind() ActionKind {
	return ActionKindInlineMenu
}

func (t inlineMenuButton) Result() string {
	return t.inlineMenu
}

// --------------------------------------------

// stateButton is a action that is used to switch to a state.
type stateButton struct {
	baseButton
	state string
}

func (t stateButton) Kind() ActionKind {
	return ActionKindState
}

func (t stateButton) Result() string {
	return t.state
}

// --------------------------------------------

type Action interface {
	Kind() ActionKind
	Result() string
}

type staticActionBuilder struct {
	buttons  []Action
	commands []Action
}

// NewActionBuilder creates a new staticActionBuilder.
func NewActionBuilder() *staticActionBuilder {
	return &staticActionBuilder{}
}

// AddTextButton adds a text action to the staticActionBuilder.
func (b *staticActionBuilder) AddTextButton(button, text string, opts ...ButtonOptions) *staticActionBuilder {
	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: text,
	})

	return b
}

// AddInlineMenuButton adds an inline menu action to the staticActionBuilder.
func (b *staticActionBuilder) AddInlineMenuButton(button, inlineMenu string, opts ...ButtonOptions) *staticActionBuilder {
	b.buttons = append(b.buttons, inlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		inlineMenu: inlineMenu,
	})

	return b
}

// AddStateButton adds a state action to the staticActionBuilder.
func (b *staticActionBuilder) AddStateButton(button, state string, opts ...ButtonOptions) *staticActionBuilder {
	b.buttons = append(b.buttons, stateButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		}, state: state,
	})

	return b
}

// AddTextCommand adds a text command to the staticActionBuilder.
func (b *staticActionBuilder) AddTextCommand(command, text string) *staticActionBuilder {
	b.commands = append(b.commands, textCommand{
		baseCommand: baseCommand{
			command: command,
		},
		text: text,
	})

	return b
}

// AddInlineMenuCommand adds an inline menu command to the staticActionBuilder.
func (b *staticActionBuilder) AddInlineMenuCommand(command, inlineMenu string) *staticActionBuilder {
	b.commands = append(b.commands, inlineMenuCommand{
		baseCommand: baseCommand{
			command: command,
		},
		inlineMenu: inlineMenu,
	})

	return b
}

// AddStateCommand adds a state command to the staticActionBuilder.
func (b *staticActionBuilder) AddStateCommand(command, state string) *staticActionBuilder {
	b.commands = append(b.commands, stateCommand{
		baseCommand: baseCommand{
			command: command,
		},
		state: state,
	})

	return b
}

// AddCustomButton adds a custom action of button type to the staticActionBuilder.
func (b *staticActionBuilder) AddCustomButton(action Action) *staticActionBuilder {
	b.buttons = append(b.buttons, action)
	return b
}

// AddCustomCommand adds a custom action of command type to the staticActionBuilder.
func (b *staticActionBuilder) AddCustomCommand(action Action) *staticActionBuilder {
	b.commands = append(b.commands, action)
	return b
}

// getButtonByButton returns the action by the button.
func (b *staticActionBuilder) getButtonByButton(button string) Action {
	for _, btn := range b.buttons {
		if btn.Result() == button {
			return btn
		}
	}

	return nil
}
