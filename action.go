package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

// Action is the single result a handler returns, replacing the old
// (SwitchAction, ShouldPass) pair. It is a sealed interface: construct values
// with Next, Stop, Error, Show, Edit, ctx.GoTo, ctx.GoToWith, ctx.ReplyText.
type Action interface {
	isAction()
}

type actionKind int

const (
	actionKindNext actionKind = iota
	actionKindStop
	actionKindError
	actionKindState
	actionKindInlineMenu
)

type actionResult struct {
	kind actionKind

	err error

	// state transition
	state string
	data  any // *D carried by GoToWith

	// inline menu
	menu string
	edit bool
}

func (actionResult) isAction() {}

// Next passes processing to the next handler (middleware chain) or falls
// through to the default handler (menu handlers).
func Next() Action {
	return actionResult{kind: actionKindNext}
}

// Stop marks the update as handled and stops processing.
func Stop() Action {
	return actionResult{kind: actionKindStop}
}

// Error routes err to the engine's error handler and stops processing.
func Error(err error) Action {
	return actionResult{kind: actionKindError, err: err}
}

// Show sends the given inline menu as a new message.
func Show(menu InlineMenuRef) Action {
	return actionResult{kind: actionKindInlineMenu, menu: menu.name}
}

// Edit edits the current message into the given inline menu.
// Only valid while processing a callback query.
func Edit(menu InlineMenuRef) Action {
	return actionResult{kind: actionKindInlineMenu, menu: menu.name, edit: true}
}

// Handler is a middleware or button hook: it receives the request context and
// returns an Action. Returning Next continues the chain; anything else stops.
type Handler func(ctx *Ctx) Action

// PanicHandler is called when a panic is recovered while processing an update.
type PanicHandler func(
	client *tgbotapi.TelegramBot,
	update tgbotapi.Update,
	err interface{},
	trace string,
)
