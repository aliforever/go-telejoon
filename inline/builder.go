package inline

import (
	"github.com/aliforever/go-telejoon"
)

// Builder collects inline buttons and converts them to an InlineActionBuilder.
type Builder struct {
	buttons   []*Button
	formation []int
	maxPerRow int
}

// Build creates a new Builder with the given buttons.
//
// Example:
//
//	inline.Build(
//	    inline.URL(text.S("Website"), "https://..."),
//	    inline.Alert(text.S("Info"), "Coming soon"),
//	).MaxPerRow(2)
func Build(btns ...*Button) *Builder {
	return &Builder{
		buttons: btns,
	}
}

// NewBuilder creates an empty Builder.
func NewBuilder() *Builder {
	return &Builder{
		buttons: []*Button{},
	}
}

// Add adds buttons to the builder.
func (b *Builder) Add(btns ...*Button) *Builder {
	b.buttons = append(b.buttons, btns...)
	return b
}

// Formation sets the button formation.
func (b *Builder) Formation(f ...int) *Builder {
	b.formation = f
	return b
}

// MaxPerRow sets the maximum number of buttons per row.
func (b *Builder) MaxPerRow(n int) *Builder {
	b.maxPerRow = n
	return b
}

// Build converts the button collection to a telejoon.InlineActionBuilder.
func (b *Builder) Build() *telejoon.InlineActionBuilder {
	iab := telejoon.NewInlineActionBuilder()

	for _, btn := range b.buttons {
		// Skip if static condition is false
		if btn.staticCond != nil && !*btn.staticCond {
			continue
		}

		b.addButton(iab, btn)
	}

	if len(b.formation) > 0 {
		iab.SetButtonFormation(b.formation...)
	}

	if b.maxPerRow > 0 {
		iab.SetMaxButtonPerRow(b.maxPerRow)
	}

	return iab
}

func (b *Builder) addButton(iab *telejoon.InlineActionBuilder, btn *Button) {
	label := btn.label.Builder()

	// Get button options
	var opts []*telejoon.ButtonOptions
	if btn.newRow || btn.alone {
		opts = append(opts, telejoon.NewButtonOptions(btn.newRow || btn.alone, btn.alone))
	}

	// Determine data
	var data telejoon.TextBuilder
	if btn.data != "" {
		data = telejoon.NewStaticText(btn.data)
	} else if btn.dataFn != nil {
		data = telejoon.NewDeferredText(btn.dataFn)
	} else {
		// Use target as default data
		data = telejoon.NewStaticText(btn.target)
	}

	switch btn.action {
	case actionURL:
		iab.AddUrlButton(label, telejoon.NewStaticText(btn.target), opts...)

	case actionAlert:
		iab.AddAlertButton(label, data, btn.target, opts...)

	case actionConfirm:
		iab.AddAlertButtonWithDialog(label, data, btn.target, opts...)

	case actionMenu:
		iab.AddInlineMenuButton(label, data, btn.target, opts...)

	case actionMenuEdit:
		iab.AddInlineMenuButtonWithEdit(label, data, btn.target, opts...)

	case actionState:
		iab.AddStateButton(label, data, btn.target, opts...)

	case actionCallback:
		iab.AddCallbackButton(label, data, btn.handler, opts...)
	}
}

// BuildInline implements InlineActionBuilderKind interface.
func (b *Builder) BuildInline(update *telejoon.StateUpdate) *telejoon.InlineActionBuilder {
	return b.Build()
}
