package telejoon

type baseButton struct {
	button TextBuilder

	definedCondition *string

	definedConditionFalse *string

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
	cond1Vs := t.definedConditionFalse == nil || !definedConditionResults[*t.definedConditionFalse]

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

// AddDefinedConditionalTextButton adds a textHandler action to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddDefinedConditionalTextButton(
	cond string,
	button TextBuilder,
	text TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:           button,
			definedCondition: &cond,
			options:          opts,
		},
		text: text,
	})

	return b
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

	b.buttons = append(b.buttons, textButton{
		baseButton: baseButton{
			button:                button,
			definedConditionFalse: &vsCond,
			options:               opts,
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

// AddConditionalRawButton adds a raw button to the ActionBuilder with a condition.
func (b *ActionBuilder) AddConditionalRawButton(
	cond func(update *StateUpdate) bool,
	button TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, rawButton{
		baseButton: baseButton{
			button:    button,
			options:   opts,
			condition: cond,
		},
	})

	return b
}

// AddDefinedConditionalRawButton adds a raw button to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddDefinedConditionalRawButton(
	cond string,
	button TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, rawButton{
		baseButton: baseButton{
			button:           button,
			options:          opts,
			definedCondition: &cond,
		},
	})

	return b
}

// AddVsDefinedConditionalRawButton adds a raw button to the ActionBuilder with a condition name.
func (b *ActionBuilder) AddVsDefinedConditionalRawButton(
	cond string,
	button TextBuilder,
	opts ...*ButtonOptions,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.buttons = append(b.buttons, rawButton{
		baseButton: baseButton{
			button:                button,
			options:               opts,
			definedConditionFalse: &cond,
		},
	})

	return b
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

	b.buttons = append(b.buttons, stateButton{
		baseButton: baseButton{
			button:    button,
			options:   opts,
			condition: cond,
		}, state: state,
	})

	return b
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

	b.buttons = append(b.buttons, stateButton{
		baseButton: baseButton{
			button:           button,
			options:          opts,
			definedCondition: &vsCond,
		}, state: state,
	})

	return b
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

	b.buttons = append(b.buttons, stateButton{
		baseButton: baseButton{
			button:                button,
			options:               opts,
			definedConditionFalse: &vsCond,
		}, state: state,
	})

	return b
}
