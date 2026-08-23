package telejoon

import "encoding/json"

// StateDataRepository is an optional extension that persists state payloads
// (carried by GoToWith transitions) alongside the state name, encoded as
// JSON. Without it, payloads live in-process only: after a restart, menus
// receive the zero value of their D.
type StateDataRepository interface {
	SetUserStateData(userID int64, state string, data []byte) error
	GetUserStateData(userID int64, state string) ([]byte, error)
}

func unmarshalStateData[D any](raw []byte, data *D) error {
	return json.Unmarshal(raw, data)
}

// NoData marks a state (or inline menu) that carries no payload.
type NoData struct{}

// State is a typed handle to a bot state. The name is what gets persisted in
// the UserRepository (so stored sessions keep working), while the type
// parameter D declares the payload a transition into this state carries.
//
// Declare states as package variables:
//
//	var (
//		Welcome  = telejoon.NewState[telejoon.NoData]("Welcome")
//		Checkout = telejoon.NewState[CheckoutData]("Checkout")
//	)
type State[D any] struct {
	name string
}

// NewState creates a new typed state handle with the given persisted name.
func NewState[D any](name string) State[D] {
	return State[D]{name: name}
}

// Name returns the persisted name of the state.
func (s State[D]) Name() string {
	return s.name
}

// StateData returns the payload of the current state, typed as D.
// It reports false when the current state does not match the given handle or
// no payload is available. Useful inside Text builders and conditions, which
// are state-agnostic and therefore do not receive data *D directly.
func StateData[D any](ctx *Ctx, s State[D]) (D, bool) {
	var zero D

	if ctx.State != s.name || ctx.stateData == nil {
		return zero, false
	}

	if data, ok := ctx.stateData.(*D); ok && data != nil {
		return *data, true
	}

	return zero, false
}

// InlineMenuRef is a handle to a registered inline menu.
type InlineMenuRef struct {
	name string
}

// NewInlineMenuRef creates a new inline menu handle with the given name.
func NewInlineMenuRef(name string) InlineMenuRef {
	return InlineMenuRef{name: name}
}

// Name returns the name of the inline menu.
func (r InlineMenuRef) Name() string {
	return r.name
}
