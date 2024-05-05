package telejoon

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
