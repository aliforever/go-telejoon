// Package buttons provides a fluent API for creating reply keyboard buttons.
package buttons

import (
	"github.com/aliforever/go-telejoon"
	"github.com/aliforever/go-telejoon/text"
)

// actionType defines what happens when the button is pressed.
type actionType int

const (
	actionGoTo actionType = iota
	actionReply
	actionShow
	actionRaw
)

// Button represents a reply keyboard button with its configuration.
type Button struct {
	label  text.Text
	action actionType
	target string // state name, menu name, or response text

	// Hook
	hook telejoon.UpdateHandler

	// Conditions
	staticCond  *bool
	dynamicCond func(*telejoon.StateUpdate) bool
	definedCond string
	inverseCond bool

	// Layout
	newRow bool
	alone  bool
}

// GoTo creates a button that switches to a state.
//
// Example:
//
//	buttons.GoTo(text.L("Nav.Home"), "Welcome")
//	buttons.GoTo(text.S("Admin"), "Admin").When(isAdmin)
func GoTo(label text.Text, state string) *Button {
	return &Button{
		label:  label,
		action: actionGoTo,
		target: state,
	}
}

// Reply creates a button that sends a text response when pressed.
//
// Example:
//
//	buttons.Reply(text.S("Help"), "Here is how to use the bot...")
func Reply(label text.Text, response string) *Button {
	return &Button{
		label:  label,
		action: actionReply,
		target: response,
	}
}

// Show creates a button that opens an inline menu.
//
// Example:
//
//	buttons.Show(text.S("📦 Products"), "ProductsMenu")
func Show(label text.Text, inlineMenu string) *Button {
	return &Button{
		label:  label,
		action: actionShow,
		target: inlineMenu,
	}
}

// Raw creates a button with no automatic action.
// Use this for buttons handled by dynamic handlers.
//
// Example:
//
//	buttons.Raw(text.S("📸 Send Photo"))
func Raw(label text.Text) *Button {
	return &Button{
		label:  label,
		action: actionRaw,
		target: "",
	}
}
