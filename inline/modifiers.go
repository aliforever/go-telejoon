package inline

import "github.com/aliforever/go-telejoon"

// === Data Modifiers ===

// Data sets the callback data for the button.
//
// Example:
//
//	inline.Callback(text.S("Delete"), handler).Data("delete:123")
func (b *Button) Data(data string) *Button {
	b.data = data
	return b
}

// DataD sets a dynamic callback data computed at render time.
//
// Example:
//
//	inline.Callback(text.S("Delete"), handler).DataD(func(u *telejoon.StateUpdate) string {
//	    return fmt.Sprintf("delete:%d", u.Get("itemId"))
//	})
func (b *Button) DataD(fn func(*telejoon.StateUpdate) string) *Button {
	b.dataFn = fn
	return b
}

// === Condition Modifiers ===

// If sets a static (compile-time) condition.
// Button is only included if cond is true.
func (b *Button) If(cond bool) *Button {
	b.staticCond = &cond
	return b
}

// When sets a dynamic (render-time) condition.
// Button visibility is evaluated per-request.
func (b *Button) When(cond func(*telejoon.StateUpdate) bool) *Button {
	b.dynamicCond = cond
	return b
}

// === Layout Modifiers ===

// NewRow places this button on a new row.
func (b *Button) NewRow() *Button {
	b.newRow = true
	return b
}

// Alone places this button on its own row.
func (b *Button) Alone() *Button {
	b.alone = true
	return b
}
