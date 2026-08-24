package telejoon

import "sync/atomic"

// Middleware semantics
//
// Engine middlewares (Engine.Use) run exactly once per update, before any
// menu processing. A menu's middleware chain (MenuBuilder.Use,
// InlineMenuBuilder.Use) runs when the menu is ENTERED — both when the
// incoming update is dispatched to it and when it is rendered as the
// destination of a state switch — but at most once per menu per update:
// re-entering the same menu within one update (e.g. a route returning
// Edit(sameMenu), or a chained redirect) does not re-run its chain.
//
// Two consequences worth knowing:
//
//   - A state switch runs the destination menu's chain too, because the
//     destination menu is entered to be rendered. Middleware that prepares
//     data for the menu's text/keyboard belongs here and keeps working.
//   - A middleware shared by several menus runs once per entered menu. Wrap
//     it with Once to run it at most once per update across all menus, or
//     with DispatchOnly to skip it on switch-render passes.

var onceCounter atomic.Uint64

// Once wraps a middleware so it runs at most once per update, no matter how
// many menus it is attached to. Use it for expensive middlewares (database
// lookups, subscription checks) that are shared across menus: a state switch
// entering several such menus in one update invokes the wrapped middleware
// only for the first.
//
// Each call to Once creates a distinct wrapper with its own identity:
//
//	check := telejoon.Once(loadProfile) // share this value across menus
//	welcome.Use(check)
//	profile.Use(check)
func Once(middleware Handler) Handler {
	id := onceCounter.Add(1)

	return func(ctx *Ctx) Action {
		if ctx.ranOnce == nil {
			ctx.ranOnce = map[uint64]struct{}{}
		}

		if _, ok := ctx.ranOnce[id]; ok {
			return Next()
		}

		ctx.ranOnce[id] = struct{}{}

		return middleware(ctx)
	}
}

// DispatchOnly wraps a middleware so it runs only when the menu dispatches
// the incoming update, and is skipped when the menu is merely rendered as
// the destination of a state switch (Ctx.IsSwitched). Use it for middleware
// whose work only matters when the user actually interacts with the menu —
// rate limiting, typing indicators, analytics.
func DispatchOnly(middleware Handler) Handler {
	return func(ctx *Ctx) Action {
		if ctx.IsSwitched {
			return Next()
		}

		return middleware(ctx)
	}
}
