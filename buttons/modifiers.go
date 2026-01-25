package buttons

import "github.com/aliforever/go-telejoon"

// === Condition Modifiers ===

// If sets a static (compile-time) condition.
// Button is only included if cond is true.
// Useful for feature flags or build configurations.
//
// Example:
//
//	buttons.GoTo(text.S("Debug"), "Debug").If(config.DebugMode)
func (b *Button) If(cond bool) *Button {
	b.staticCond = &cond
	return b
}

// When sets a dynamic (render-time) condition.
// Button visibility is evaluated per-request.
//
// Example:
//
//	buttons.GoTo(text.S("Admin"), "Admin").When(func(u *telejoon.StateUpdate) bool {
//	    return u.Get("role") == "admin"
//	})
func (b *Button) When(cond func(*telejoon.StateUpdate) bool) *Button {
	b.dynamicCond = cond
	return b
}

// WhenDefined uses a named condition from the builder.
// The condition must be defined with Builder.Define().
//
// Example:
//
//	builder.Define("isAdmin", isAdminFunc)
//	buttons.GoTo(text.S("Admin"), "Admin").WhenDefined("isAdmin")
func (b *Button) WhenDefined(name string) *Button {
	b.definedCond = name
	b.inverseCond = false
	return b
}

// Unless is inverse of When - button shown when condition is false.
//
// Example:
//
//	buttons.GoTo(text.S("Login"), "Login").Unless(isLoggedIn)
func (b *Button) Unless(cond func(*telejoon.StateUpdate) bool) *Button {
	b.dynamicCond = cond
	b.inverseCond = true
	return b
}

// UnlessDefined is inverse of WhenDefined.
// Button is shown when the named condition is false.
//
// Example:
//
//	buttons.GoTo(text.S("Upgrade"), "Upgrade").UnlessDefined("isPremium")
func (b *Button) UnlessDefined(name string) *Button {
	b.definedCond = name
	b.inverseCond = true
	return b
}

// === Layout Modifiers ===

// NewRow places this button on a new row.
//
// Example:
//
//	buttons.GoTo(text.S("Back"), "Main").NewRow()
func (b *Button) NewRow() *Button {
	b.newRow = true
	return b
}

// Alone places this button on its own row (break before and after).
//
// Example:
//
//	buttons.GoTo(text.S("Submit"), "Submit").Alone()
func (b *Button) Alone() *Button {
	b.alone = true
	return b
}

// === Hook Modifiers ===

// Before runs a handler before the action executes.
// Use for logging, validation, or setting context values.
//
// Example:
//
//	buttons.GoTo(text.S("Products"), "Products").Before(logAccess)
func (b *Button) Before(hook telejoon.UpdateHandler) *Button {
	b.hook = hook
	return b 
}
