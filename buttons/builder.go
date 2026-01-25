package buttons

import (
	"github.com/aliforever/go-telejoon"
)

// Builder collects buttons and converts them to an ActionBuilder.
type Builder struct {
	buttons    []*Button
	conditions map[string]func(*telejoon.StateUpdate) bool
	formation  []int
	maxPerRow  int
}

// Build creates a new Builder with the given buttons.
//
// Example:
//
//	buttons.Build(
//	    buttons.GoTo(text.L("Nav.Home"), "Welcome"),
//	    buttons.GoTo(text.S("Settings"), "Settings"),
//	).Formation(2)
func Build(btns ...*Button) *Builder {
	return &Builder{
		buttons:    btns,
		conditions: make(map[string]func(*telejoon.StateUpdate) bool),
	}
}

// NewBuilder creates an empty Builder.
func NewBuilder() *Builder {
	return &Builder{
		buttons:    []*Button{},
		conditions: make(map[string]func(*telejoon.StateUpdate) bool),
	}
}

// Add adds buttons to the builder.
func (b *Builder) Add(btns ...*Button) *Builder {
	b.buttons = append(b.buttons, btns...)
	return b
}

// Define defines a named condition for use with WhenDefined/UnlessDefined.
//
// Example:
//
//	builder.Define("isAdmin", func(u *telejoon.StateUpdate) bool {
//	    return u.Get("role") == "admin"
//	})
func (b *Builder) Define(name string, cond func(*telejoon.StateUpdate) bool) *Builder {
	b.conditions[name] = cond
	return b
}

// Formation sets the button formation (buttons per row).
//
// Example:
//
//	builder.Formation(2, 1, 1)  // 2 buttons, then 1, then 1
func (b *Builder) Formation(f ...int) *Builder {
	b.formation = f
	return b
}

// MaxPerRow sets the maximum number of buttons per row.
//
// Example:
//
//	builder.MaxPerRow(3)
func (b *Builder) MaxPerRow(n int) *Builder {
	b.maxPerRow = n
	return b
}

// Build converts the button collection to a telejoon.ActionBuilder.
func (b *Builder) Build() *telejoon.ActionBuilder {
	ab := telejoon.NewStaticActionBuilder()

	// Register defined conditions
	for name, cond := range b.conditions {
		ab.DefineCondition(name, cond)
	}

	// Convert buttons to actions
	for _, btn := range b.buttons {
		// Skip if static condition is false
		if btn.staticCond != nil && !*btn.staticCond {
			continue
		}

		action := btn.toAction()
		if action != nil {
			ab.AddCustomButton(action)
		}
	}

	// Set formation
	if len(b.formation) > 0 {
		ab.SetButtonFormation(b.formation...)
	}

	if b.maxPerRow > 0 {
		ab.SetMaxButtonPerRow(b.maxPerRow)
	}

	return ab
}

// toAction converts a Button to a telejoon.Action.
func (btn *Button) toAction() telejoon.Action {
	// Build button options
	var opts []*telejoon.ButtonOptions
	if btn.newRow || btn.alone {
		opts = append(opts, telejoon.NewButtonOptions(btn.newRow || btn.alone, btn.alone))
	}

	labelBuilder := btn.label.Builder()

	switch btn.action {
	case actionGoTo:
		if btn.dynamicCond != nil {
			if btn.inverseCond {
				// No inverse conditional state button in current API, use regular with wrapper
				return telejoon.ConditionalStateButton(
					func(u *telejoon.StateUpdate) bool { return !btn.dynamicCond(u) },
					labelBuilder, btn.target, opts...)
			}
			return telejoon.ConditionalStateButton(btn.dynamicCond, labelBuilder, btn.target, opts...)
		}
		if btn.definedCond != "" {
			if btn.inverseCond {
				return telejoon.VsDefinedConditionalStateButton(btn.definedCond, labelBuilder, btn.target, opts...)
			}
			return telejoon.DefinedConditionalStateButton(btn.definedCond, labelBuilder, btn.target, opts...)
		}
		if btn.hook != nil {
			return telejoon.StateButtonWithHook(labelBuilder, btn.target, btn.hook, opts...)
		}
		return telejoon.StateButton(labelBuilder, btn.target, opts...)

	case actionReply:
		responseBuilder := telejoon.NewStaticText(btn.target)
		if btn.dynamicCond != nil {
			if btn.inverseCond {
				return telejoon.ConditionalTextButton(
					func(u *telejoon.StateUpdate) bool { return !btn.dynamicCond(u) },
					labelBuilder, responseBuilder, opts...)
			}
			return telejoon.ConditionalTextButton(btn.dynamicCond, labelBuilder, responseBuilder, opts...)
		}
		if btn.definedCond != "" {
			if btn.inverseCond {
				return telejoon.VsDefinedConditionalTextButton(btn.definedCond, labelBuilder, responseBuilder, opts...)
			}
			return telejoon.DefinedConditionalTextButton(btn.definedCond, labelBuilder, responseBuilder, opts...)
		}
		return telejoon.TextButton(labelBuilder, responseBuilder, opts...)

	case actionShow:
		return telejoon.InlineMenuButton(labelBuilder, btn.target, opts...)

	case actionRaw:
		if btn.dynamicCond != nil {
			if btn.inverseCond {
				return telejoon.ConditionalRawButton(
					func(u *telejoon.StateUpdate) bool { return !btn.dynamicCond(u) },
					labelBuilder, opts...)
			}
			return telejoon.ConditionalRawButton(btn.dynamicCond, labelBuilder, opts...)
		}
		if btn.definedCond != "" {
			if btn.inverseCond {
				return telejoon.VsDefinedConditionalRawButton(btn.definedCond, labelBuilder, opts...)
			}
			return telejoon.DefinedConditionalRawButton(btn.definedCond, labelBuilder, opts...)
		}
		return telejoon.RawButton(labelBuilder, opts...)
	}

	return nil
}

// build implements ActionBuilderKind interface.
func (b *Builder) build(update *telejoon.StateUpdate) *telejoon.ActionBuilder {
	return b.Build().
		// Re-evaluate conditions at build time
		SetConditionValue("", false) // Trigger condition evaluation
}
