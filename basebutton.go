package telejoon

type baseButton struct {
	button TextBuilder

	definedCondition *string

	vsDefinedCondition *string

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

func (t baseButton) CanBeShown(update *StateUpdate, definedConditionResults map[string]bool) bool {
	cond1 := t.definedCondition == nil || definedConditionResults[*t.definedCondition]
	cond1Vs := t.vsDefinedCondition == nil || !definedConditionResults[*t.vsDefinedCondition]

	cond2 := t.condition == nil || t.condition(update)

	return (cond1 && cond1Vs) && cond2
}

type baseButtonOptions interface {
	Options() *ButtonOptions
	CanBeShown(update *StateUpdate, definedConditionResults map[string]bool) bool
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

func TextButton(button TextBuilder, text TextBuilder, opts ...*ButtonOptions) Action {
	return textButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		text: text,
	}
}

// AddTextButton adds a textHandler action to the ActionBuilder.
func (b *ActionBuilder) AddTextButton(button TextBuilder, text TextBuilder, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, TextButton(button, text, opts...))

	return b
}

func ConditionalTextButton(
	cond func(update *StateUpdate) bool,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) Action {
	return textButton{
		baseButton: baseButton{
			button:    button,
			condition: cond,
			options:   opts,
		},
		text: text,
	}
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

	b.buttons = append(b.buttons, ConditionalTextButton(cond, button, text, opts...))

	return b
}

func DefinedConditionalTextButton(
	cond string,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) Action {
	return textButton{
		baseButton: baseButton{
			button:           button,
			definedCondition: &cond,
			options:          opts,
		},
		text: text,
	}
}

// AddDefinedConditionalTextButton adds a textHandler action to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddDefinedConditionalTextButton(
	cond string,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, DefinedConditionalTextButton(cond, button, text, opts...))

	return b
}

func VsDefinedConditionalTextButton(
	vsCond string,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) Action {
	return textButton{
		baseButton: baseButton{
			button:             button,
			vsDefinedCondition: &vsCond,
			options:            opts,
		},
		text: text,
	}
}

// AddVsDefinedConditionalTextButton adds a textHandler action to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddVsDefinedConditionalTextButton(
	vsCond string,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, VsDefinedConditionalTextButton(vsCond, button, text, opts...))

	return b
}

func InlineMenuButton(button TextBuilder, inlineMenu string, opts ...*ButtonOptions) Action {
	return inlineMenuButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		inlineMenu: inlineMenu,
	}
}

// AddInlineMenuButton adds an inline menu action to the ActionBuilder.
func (b *ActionBuilder) AddInlineMenuButton(
	button TextBuilder,
	inlineMenu string,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, InlineMenuButton(button, inlineMenu, opts...))

	return b
}

func StateButton(button TextBuilder, state string, opts ...*ButtonOptions) Action {
	return stateButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		state: state,
	}
}

// AddStateButton adds a state action to the ActionBuilder.
func (b *ActionBuilder) AddStateButton(button TextBuilder, state string, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, StateButton(button, state, opts...))

	return b
}

func StateButtonWithHook(button TextBuilder, state string, hook UpdateHandler, opts ...*ButtonOptions) Action {
	return stateButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
		state: state,
		hook:  hook,
	}
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

	b.buttons = append(b.buttons, StateButtonWithHook(button, state, hook, opts...))

	return b
}

func RawButton(button TextBuilder, opts ...*ButtonOptions) Action {
	return rawButton{
		baseButton: baseButton{
			button:  button,
			options: opts,
		},
	}
}

// AddRawButton adds a raw button to the ActionBuilder.
func (b *ActionBuilder) AddRawButton(button TextBuilder, opts ...*ButtonOptions) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, RawButton(button, opts...))

	return b
}

func ConditionalRawButton(cond func(update *StateUpdate) bool, button TextBuilder, opts ...*ButtonOptions) Action {
	return rawButton{
		baseButton: baseButton{
			button:    button,
			options:   opts,
			condition: cond,
		},
	}
}

// AddConditionalRawButton adds a raw button to the ActionBuilder with a condition.
func (b *ActionBuilder) AddConditionalRawButton(
	cond func(update *StateUpdate) bool,
	button TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, ConditionalRawButton(cond, button, opts...))

	return b
}

func DefinedConditionalRawButton(cond string, button TextBuilder, opts ...*ButtonOptions) Action {
	return rawButton{
		baseButton: baseButton{
			button:           button,
			options:          opts,
			definedCondition: &cond,
		},
	}
}

// AddDefinedConditionalRawButton adds a raw button to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddDefinedConditionalRawButton(
	cond string,
	button TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, DefinedConditionalRawButton(cond, button, opts...))

	return b
}

func VsDefinedConditionalRawButton(cond string, button TextBuilder, opts ...*ButtonOptions) Action {
	return rawButton{
		baseButton: baseButton{
			button:             button,
			options:            opts,
			vsDefinedCondition: &cond,
		},
	}
}

// AddVsDefinedConditionalRawButton adds a raw button to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddVsDefinedConditionalRawButton(
	cond string,
	button TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, VsDefinedConditionalRawButton(cond, button, opts...))

	return b
}

func ConditionalStateButton(
	cond func(update *StateUpdate) bool,
	button TextBuilder,
	state string,
	opts ...*ButtonOptions,
) Action {

	return stateButton{
		baseButton: baseButton{
			button:    button,
			options:   opts,
			condition: cond,
		}, state: state,
	}
}

// AddConditionalStateButton adds a state button to the ActionBuilder with a condition.
func (b *ActionBuilder) AddConditionalStateButton(
	cond func(update *StateUpdate) bool,
	button TextBuilder,
	state string,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, ConditionalStateButton(cond, button, state, opts...))

	return b
}

func DefinedConditionalStateButton(
	vsCond string,
	button TextBuilder,
	state string,
	opts ...*ButtonOptions,
) Action {

	return stateButton{
		baseButton: baseButton{
			button:           button,
			options:          opts,
			definedCondition: &vsCond,
		}, state: state,
	}
}

// AddDefinedConditionalStateButton adds a state button to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddDefinedConditionalStateButton(
	vsCond string,
	button TextBuilder,
	state string,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, DefinedConditionalStateButton(vsCond, button, state, opts...))

	return b
}

func VsDefinedConditionalStateButton(
	vsCond string,
	button TextBuilder,
	state string,
	opts ...*ButtonOptions,
) Action {

	return stateButton{
		baseButton: baseButton{
			button:             button,
			options:            opts,
			vsDefinedCondition: &vsCond,
		}, state: state,
	}
}

// AddVsDefinedConditionalStateButton adds a state button to the ActionBuilder with a vs condition name
func (b *ActionBuilder) AddVsDefinedConditionalStateButton(
	vsCond string,
	button TextBuilder,
	state string,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, VsDefinedConditionalStateButton(vsCond, button, state, opts...))

	return b
}
