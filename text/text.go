// Package text provides a type-safe Text type for button labels.
package text

import "github.com/aliforever/go-telejoon"

// Text wraps a TextBuilder for type-safe label passing to button constructors.
type Text struct {
	builder telejoon.TextBuilder
}

// Builder returns the underlying TextBuilder.
func (t Text) Builder() telejoon.TextBuilder {
	return t.builder
}

// S creates a static text.
// Use for plain text labels and responses.
//
// Example:
//
//	buttons.GoTo(text.S("Home"), "Welcome")
func S(s string) Text {
	return Text{builder: telejoon.NewStaticText(s)}
}

// L creates a language key text (localized).
// Use for labels that will be translated based on user language.
//
// Example:
//
//	buttons.GoTo(text.L("Nav.Home"), "Welcome")
func L(key string) Text {
	return Text{builder: telejoon.NewLanguageKeyText(key)}
}

// D creates a deferred text (computed at render time).
// Use for dynamic labels that depend on user state.
//
// Example:
//
//	buttons.GoTo(text.D(func(u *telejoon.StateUpdate) string {
//	    return fmt.Sprintf("Cart (%d)", u.Get("count"))
//	}), "Cart")
func D(fn func(*telejoon.StateUpdate) string) Text {
	return Text{builder: telejoon.NewDeferredText(fn)}
}

// K creates an update key text (from StateUpdate storage).
// Use for text stored in the StateUpdate context.
//
// Example:
//
//	// Assuming u.Set("greeting", "Hello!") was called
//	buttons.Reply(text.K("greeting"), "response")
func K(key string) Text {
	return Text{builder: telejoon.NewUpdateKeyText(key)}
}
