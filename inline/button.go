// Package inline provides a fluent API for creating inline keyboard buttons.
package inline

import (
	"github.com/aliforever/go-telejoon"
	"github.com/aliforever/go-telejoon/text"
)

// actionType defines the inline button action.
type actionType int

const (
	actionURL actionType = iota
	actionAlert
	actionConfirm
	actionMenu
	actionMenuEdit
	actionState
	actionCallback
)

// Button represents an inline keyboard button with its configuration.
type Button struct {
	label  text.Text
	action actionType
	target string // URL, alert text, menu name, state name

	// Callback
	handler telejoon.CallbackHandler
	data    string
	dataFn  func(*telejoon.StateUpdate) string

	// Conditions
	staticCond  *bool
	dynamicCond func(*telejoon.StateUpdate) bool

	// Layout
	newRow bool
	alone  bool
}

// URL creates a button that opens a URL.
//
// Example:
//
//	inline.URL(text.S("🌐 Website"), "https://example.com")
func URL(label text.Text, url string) *Button {
	return &Button{
		label:  label,
		action: actionURL,
		target: url,
	}
}

// Alert creates a button that shows a toast notification.
//
// Example:
//
//	inline.Alert(text.S("🔔 Notify"), "Feature coming soon!")
func Alert(label text.Text, alertText string) *Button {
	return &Button{
		label:  label,
		action: actionAlert,
		target: alertText,
	}
}

// Confirm creates a button that shows a popup dialog.
//
// Example:
//
//	inline.Confirm(text.S("🗑️ Delete"), "Are you sure?")
func Confirm(label text.Text, confirmText string) *Button {
	return &Button{
		label:  label,
		action: actionConfirm,
		target: confirmText,
	}
}

// Menu creates a button that navigates to an inline menu (new message).
//
// Example:
//
//	inline.Menu(text.S("Settings"), "SettingsMenu")
func Menu(label text.Text, menuName string) *Button {
	return &Button{
		label:  label,
		action: actionMenu,
		target: menuName,
	}
}

// MenuEdit creates a button that edits current message to an inline menu.
//
// Example:
//
//	inline.MenuEdit(text.L("Nav.Back"), "MainMenu")
func MenuEdit(label text.Text, menuName string) *Button {
	return &Button{
		label:  label,
		action: actionMenuEdit,
		target: menuName,
	}
}

// State creates a button that switches to a state.
//
// Example:
//
//	inline.State(text.S("🏠 Home"), "Welcome")
func State(label text.Text, stateName string) *Button {
	return &Button{
		label:  label,
		action: actionState,
		target: stateName,
	}
}

// Callback creates a button with a custom callback handler.
//
// Example:
//
//	inline.Callback(text.S("Delete"), deleteHandler).Data("delete:123")
func Callback(label text.Text, handler telejoon.CallbackHandler) *Button {
	return &Button{
		label:   label,
		action:  actionCallback,
		handler: handler,
	}
}
