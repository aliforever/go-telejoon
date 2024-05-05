package telejoon

import (
	"sync"

	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telegram-bot-api/tools"
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

type conditionalButtonFormation struct {
	cond      func(update *StateUpdate) bool
	formation []int
}

type ActionBuilder struct {
	locker sync.Mutex

	definedConditions map[string]func(update *StateUpdate) bool

	conditionalButtons []conditionalButtons
	buttons            []Action
	commands           []Action

	conditionalButtonFormations []conditionalButtonFormation

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

func (b *ActionBuilder) AddConditionalButtonFormation(
	cond func(update *StateUpdate) bool,
	formation ...int,
) *ActionBuilder {

	b.locker.Lock()
	defer b.locker.Unlock()

	b.conditionalButtonFormations = append(b.conditionalButtonFormations, conditionalButtonFormation{
		cond:      cond,
		formation: formation,
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

// DefineCondition defines a condition.
func (b *ActionBuilder) DefineCondition(name string, cond func(update *StateUpdate) bool) *ActionBuilder {
	b.locker.Lock()
	defer b.locker.Unlock()

	if b.definedConditions == nil {
		b.definedConditions = make(map[string]func(update *StateUpdate) bool)
	}

	b.definedConditions[name] = cond

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

	var newButtons []string

	buttonFormation := b.buttonFormation

	if len(b.conditionalButtonFormations) > 0 {
		for _, formation := range b.conditionalButtonFormations {
			if formation.cond(update) {
				buttonFormation = formation.formation
				break
			}
		}
	}

	definedConditionsResults := make(map[string]bool)

	if len(b.definedConditions) > 0 {
		for name, cond := range b.definedConditions {
			definedConditionsResults[name] = cond(update)
		}
	}

	if len(b.conditionalButtons) > 0 {
		for _, button := range b.conditionalButtons {
			availableButtons := b.makeButtonsFromActions(update, definedConditionsResults, button.buttons)
			if len(availableButtons) > 0 {
				newButtons = append(newButtons, availableButtons...)

				if len(button.formation) > 0 {
					buttonFormation = append(buttonFormation, button.formation...)
				}
			}
		}
	}

	mainButtons := b.makeButtonsFromActions(update, definedConditionsResults, b.buttons)

	newButtons = append(newButtons, mainButtons...)
	buttonFormation = append(buttonFormation, b.buttonFormation...)

	return tools.Keyboards{}.NewReplyKeyboardFromSliceOfStringsWithFormation(
		newButtons,
		b.maxButtonPerRow,
		buttonFormation,
		reverseButtonOrderInRows,
	)
}

func (b *ActionBuilder) makeButtonsFromActions(
	update *StateUpdate,
	definedConditionsResults map[string]bool,
	actions []Action,
) []string {
	var newButtons []string

	for _, button := range actions {
		name := button.Name(update)

		shouldBreakAfter := false

		if opts, ok := button.(baseButtonOptions); ok {
			if !opts.CanBeShown(update, definedConditionsResults) {
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

	return newButtons
}
