package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
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

type baseCommand struct {
	command string
}

// textCommand is a command that sends a text message.
type textCommand struct {
	baseCommand
	text string
}

func (t textCommand) Name() string {
	return t.command
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

func (t inlineMenuCommand) Name() string {
	return t.command
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

func (t stateCommand) Name() string {
	return t.command
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
	options []*ButtonOptions
}

// textButton is a action that sends a text message.
type textButton struct {
	baseButton
	text string
}

func (t textButton) Name() string {
	return t.button
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

func (t inlineMenuButton) Name() string {
	return t.button
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

func (t stateButton) Name() string {
	return t.button
}

func (t stateButton) Kind() ActionKind {
	return ActionKindState
}

func (t stateButton) Result() string {
	return t.state
}

// --------------------------------------------

// rawButton is only a raw button that does nothing but sends the button name.
type rawButton struct {
	baseButton
}

func (t rawButton) Name() string {
	return t.button
}

func (t rawButton) Kind() ActionKind {
	return ActionKindRaw
}

func (t rawButton) Result() string {
	return ""
}

// --------------------------------------------

type Action interface {
	Name() string
	Kind() ActionKind
	Result() string
}

type ActionBuilder struct {
	locker sync.Mutex

	buttons  []Action
	commands []Action

	buttonOptions map[string][]*ButtonOptions

	buttonFormation []int
	maxButtonPerRow int
}

// NewActionBuilder creates a new ActionBuilder.
func NewActionBuilder() *ActionBuilder {
	return &ActionBuilder{
		buttonOptions: make(map[string][]*ButtonOptions),
	}
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
func (b *ActionBuilder) AddTextButton(button, text string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: text,
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	return b
}

// AddTextButtonT adds a textHandler action to the ActionBuilder with name translation.
func (b *ActionBuilder) AddTextButtonT(button, text string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: text,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// AddInlineMenuButton adds an inline menu action to the ActionBuilder.
func (b *ActionBuilder) AddInlineMenuButton(
	button, inlineMenu string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		inlineMenu: inlineMenu,
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	return b
}

// AddInlineMenuButtonT adds an inline menu action to the ActionBuilder with name translation.
func (b *ActionBuilder) AddInlineMenuButtonT(button, inlineMenu string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, inlineMenuButton{
		baseButton: baseButton{
			button: button,
		},
		inlineMenu: inlineMenu,
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// AddStateButton adds a state action to the ActionBuilder.
func (b *ActionBuilder) AddStateButton(button, state string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, stateButton{
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

// AddStateButtonT adds a state action to the ActionBuilder with name translation.
func (b *ActionBuilder) AddStateButtonT(button, state string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, stateButton{
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

// AddRawButton adds a raw button to the ActionBuilder.
func (b *ActionBuilder) AddRawButton(button string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, rawButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
	})

	if len(opts) > 0 {
		b.buttonOptions[button] = opts
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// AddRawButtonT adds a raw button to the ActionBuilder with name translation.
func (b *ActionBuilder) AddRawButtonT(button string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, rawButton{
		baseButton: baseButton{
			button: button,
		},
	})

	b.buttonOptions[button] = []*ButtonOptions{
		NewButtonOptions().TranslateName(),
	}

	if len(opts) > 0 {
		b.buttonOptions[button] = append(b.buttonOptions[button], opts...)
	}

	return b
}

// AddTextCommand adds a textHandler command to the ActionBuilder.
func (b *ActionBuilder) AddTextCommand(command, text string) *ActionBuilder {
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
func (b *ActionBuilder) AddInlineMenuCommand(command, inlineMenu string) *ActionBuilder {
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
func (b *ActionBuilder) AddStateCommand(command, state string) *ActionBuilder {
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
func (b *ActionBuilder) AddCustomButton(action Action, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, action)

	if len(opts) > 0 {
		b.buttonOptions[action.Name()] = append(b.buttonOptions[action.Name()], opts...)
	}

	return b
}

// AddCustomCommand adds a custom action of command type to the ActionBuilder.
func (b *ActionBuilder) AddCustomCommand(action Action) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.commands = append(b.commands, action)

	return b
}

// getButtonByButton returns the action by the button.
func (b *ActionBuilder) getButtonByButton(button string) Action {
	for _, btn := range b.buttons {
		if btn.Name() == button {
			return btn
		}
	}

	return nil
}

// buildButtons builds the buttons.
func (b *ActionBuilder) buildButtons(language *Language) *structs.ReplyKeyboardMarkup {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(b.buttons) == 0 {
		return nil
	}

	newButtons := []string{}

	for _, button := range b.buttons {
		name := button.Name()

		shouldBreakAfter := false

		if opts := b.buttonOptions[button.Name()]; len(opts) > 0 {
			if language != nil {
				if opts[0].translateName {
					btnText, err := language.Get(button.Name())
					if err == nil {
						name = btnText
					}
				}
			}

			if opts[0].breakBefore {
				newButtons = append(newButtons, "")
			}

			if opts[0].breakAfter {
				shouldBreakAfter = true
			}
		}

		newButtons = append(newButtons, name)

		if shouldBreakAfter {
			newButtons = append(newButtons, "")
		}
	}

	return tools.Keyboards{}.NewReplyKeyboardFromSliceOfStringsWithFormation(
		newButtons, b.maxButtonPerRow, b.buttonFormation)
}

func (b *ActionBuilder) languageValueButtonKeys(language *Language) map[string]string {
	b.locker.Lock()
	defer b.locker.Unlock()

	var valueKeys = make(map[string]string)

	for k, v := range b.buttonOptions {
		if len(v) > 0 && v[0].translateName {
			keyValue, err := language.Get(k)
			if err != nil {
				valueKeys[k] = k
			} else {
				valueKeys[keyValue] = k
			}
		}
	}

	return valueKeys
}

// chooseLanguageButton implements the Action interface.
type chooseLanguageButton[User any] struct {
	button         string
	engine         *EngineWithPrivateStateHandlers[User]
	client         *tgbotapi.TelegramBot
	update         *StateUpdate[User]
	languageConfig *LanguageConfig
	localizer      *Language
}

// NewChooseLanguageButton creates a new chooseLanguageButton.
func NewChooseLanguageButton[User any](
	button string, engine *EngineWithPrivateStateHandlers[User], update *StateUpdate[User], lang *LanguageConfig,
	localizer *Language, client *tgbotapi.TelegramBot) Action {

	return chooseLanguageButton[User]{
		button:         button,
		engine:         engine,
		languageConfig: lang,
		client:         client,
		update:         update,
		localizer:      localizer,
	}
}

func (c chooseLanguageButton[User]) Name() string {
	return c.button
}

func (c chooseLanguageButton[User]) Kind() ActionKind {
	return ActionKindState
}

func (c chooseLanguageButton[User]) Result() string {
	err := c.languageConfig.repo.SetUserLanguage(c.update.Update.From().Id, c.localizer.tag)
	if err != nil {
		c.engine.onErr(c.client, c.update.Update, err)
		return ""
	}

	c.update.SetLanguage(c.localizer)

	return c.engine.defaultStateName
}
