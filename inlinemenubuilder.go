package telejoon

import "fmt"

// InlineMenuBuilder builds an inline menu. Compile it into the engine with
// engine.Add.
type InlineMenuBuilder struct {
	ref  InlineMenuRef
	text Text

	buttons   []*InlineButton
	buttonsFn func(ctx *Ctx) []*InlineButton

	routes map[string]erasedRouteHandler

	middlewares []Handler

	formation []int
	maxPerRow int
}

// InlineMenuFor creates a builder for the inline menu bound to ref.
func InlineMenuFor(ref InlineMenuRef, text Text) *InlineMenuBuilder {
	return &InlineMenuBuilder{
		ref:    ref,
		text:   text,
		routes: map[string]erasedRouteHandler{},
	}
}

// Buttons sets the static inline-keyboard buttons of the menu.
func (b *InlineMenuBuilder) Buttons(buttons ...*InlineButton) *InlineMenuBuilder {
	b.buttons = append(b.buttons, buttons...)

	return b
}

// ButtonsFunc sets a per-request inline-keyboard button factory.
func (b *InlineMenuBuilder) ButtonsFunc(fn func(ctx *Ctx) []*InlineButton) *InlineMenuBuilder {
	b.buttonsFn = fn

	return b
}

// Route registers a typed callback handler on this menu and returns the
// typed handle used to mint buttons with Do. A is inferred from the handler's
// parameter, so encoder and decoder always agree:
//
//	del := products.Route("del", func(ctx *telejoon.Ctx, a DelArgs) telejoon.Action {
//		return telejoon.Edit(Products)
//	})
func (b *InlineMenuBuilder) Route[A any](
	name string,
	fn func(ctx *Ctx, args A) Action,
	opts ...RouteOption[A],
) Route[A] {

	if !validRouteName(name) {
		panic(fmt.Sprintf("telejoon: invalid route name %q", name))
	}

	if _, exists := b.routes[name]; exists {
		panic(fmt.Sprintf("telejoon: duplicate route %q on inline menu %q", name, b.ref.name))
	}

	route := Route[A]{menu: b.ref.name, name: name, codec: positionalCodec[A]{}}
	for _, opt := range opts {
		opt(&route)
	}

	b.routes[name] = route.erase(fn)

	return route
}

// Use adds a middleware that runs before any callback of this menu is
// processed, and before the menu is rendered.
func (b *InlineMenuBuilder) Use(middleware Handler) *InlineMenuBuilder {
	b.middlewares = append(b.middlewares, middleware)

	return b
}

// Formation sets the inline-keyboard row sizes (see chunkIntoRows).
func (b *InlineMenuBuilder) Formation(formation ...int) *InlineMenuBuilder {
	b.formation = formation

	return b
}

// MaxPerRow caps the number of inline-keyboard buttons per row.
func (b *InlineMenuBuilder) MaxPerRow(max int) *InlineMenuBuilder {
	b.maxPerRow = max

	return b
}

func (b *InlineMenuBuilder) register(e *Engine) error {
	if !validRouteName(b.ref.name) {
		return fmt.Errorf("invalid_inline_menu_name: %q", b.ref.name)
	}

	if _, exists := e.inlineMenus[b.ref.name]; exists {
		return fmt.Errorf("duplicate_inline_menu: %s", b.ref.name)
	}

	if _, conflict := e.globalRoutes[b.ref.name]; conflict {
		return fmt.Errorf("inline_menu_conflicts_with_global_route: %s", b.ref.name)
	}

	e.inlineMenus[b.ref.name] = &inlineMenuRuntime{
		name:        b.ref.name,
		text:        b.text,
		buttons:     b.buttons,
		buttonsFn:   b.buttonsFn,
		routes:      b.routes,
		middlewares: b.middlewares,
		formation:   b.formation,
		maxPerRow:   b.maxPerRow,
	}

	return nil
}
