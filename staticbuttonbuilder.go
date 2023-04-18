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

type Action interface {
	Name() string
	Kind() ActionKind
	Result() string
}

type staticActionBuilder struct {
	locker sync.Mutex

	buttons  []Action
	commands []Action

	buttonOptions map[string][]*ButtonOptions
}

// NewActionBuilder creates a new staticActionBuilder.
func NewActionBuilder() *staticActionBuilder {
	return &staticActionBuilder{
		buttonOptions: make(map[string][]*ButtonOptions),
	}
}

// AddTextButton adds a text action to the staticActionBuilder.
func (b *staticActionBuilder) AddTextButton(button, text string, opts ...*ButtonOptions) *staticActionBuilder {
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

// AddTextButtonT adds a text action to the staticActionBuilder with name translation.
func (b *staticActionBuilder) AddTextButtonT(button, text string, opts ...*ButtonOptions) *staticActionBuilder {
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

	return b
}

// AddInlineMenuButton adds an inline menu action to the staticActionBuilder.
func (b *staticActionBuilder) AddInlineMenuButton(
	button, inlineMenu string, opts ...*ButtonOptions) *staticActionBuilder {
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

// AddInlineMenuButtonT adds an inline menu action to the staticActionBuilder with name translation.
func (b *staticActionBuilder) AddInlineMenuButtonT(button, inlineMenu string) *staticActionBuilder {
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

	return b
}

// AddStateButton adds a state action to the staticActionBuilder.
func (b *staticActionBuilder) AddStateButton(button, state string, opts ...*ButtonOptions) *staticActionBuilder {
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

	return b
}

// AddStateButtonT adds a state action to the staticActionBuilder with name translation.
func (b *staticActionBuilder) AddStateButtonT(button, state string) *staticActionBuilder {
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

	return b
}

// AddTextCommand adds a text command to the staticActionBuilder.
func (b *staticActionBuilder) AddTextCommand(command, text string) *staticActionBuilder {
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

// AddInlineMenuCommand adds an inline menu command to the staticActionBuilder.
func (b *staticActionBuilder) AddInlineMenuCommand(command, inlineMenu string) *staticActionBuilder {
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

// AddStateCommand adds a state command to the staticActionBuilder.
func (b *staticActionBuilder) AddStateCommand(command, state string) *staticActionBuilder {
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

// AddCustomButton adds a custom action of button type to the staticActionBuilder.
func (b *staticActionBuilder) AddCustomButton(action Action) *staticActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, action)

	return b
}

// AddCustomCommand adds a custom action of command type to the staticActionBuilder.
func (b *staticActionBuilder) AddCustomCommand(action Action) *staticActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.commands = append(b.commands, action)

	return b
}

// getButtonByButton returns the action by the button.
func (b *staticActionBuilder) getButtonByButton(button string) Action {
	for _, btn := range b.buttons {
		if btn.Name() == button {
			return btn
		}
	}

	return nil
}

// buildButtons builds the buttons.
func (b *staticActionBuilder) buildButtons(language *Language) *structs.ReplyKeyboardMarkup {
	b.locker.Lock()
	defer b.locker.Unlock()

	if len(b.buttons) == 0 {
		return nil
	}

	var newButtons = make([]string, len(b.buttons))

	for i, button := range b.buttons {
		newButtons[i] = button.Name()
	}

	if language != nil {
		for i, button := range b.buttons {
			if opts := b.buttonOptions[button.Name()]; len(opts) > 0 {
				if opts[0].translateName {
					btnText, err := language.Get(button.Name())
					if err == nil {
						newButtons[i] = btnText
					}
				}
			}
		}
	}

	return tools.Keyboards{}.NewReplyKeyboardFromSliceOfStrings(newButtons, 2)
}

func (b *staticActionBuilder) languageValueButtonKeys(language *Language) map[string]string {
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
