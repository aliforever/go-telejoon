// Command shopbot is a comprehensive telejoon v3 example: a small shop bot
// demonstrating every feature of the framework. It compiles as part of the
// module and runs with a real bot token:
//
//	BOT_TOKEN=... go run ./examples/shopbot
//
// Files: states.go (typed handles), store.go (demo data), menus.go (state
// menus), inline.go (inline menus + typed routes), main.go (wiring).
package main

import (
	"github.com/aliforever/go-telejoon"
)

// === States ===
//
// States are typed package variables. The string name is what gets persisted
// in the UserRepository; the type parameter declares the payload a transition
// into the state carries.

var (
	stateWelcome    = telejoon.NewState[telejoon.NoData]("Welcome")
	stateProducts   = telejoon.NewState[telejoon.NoData]("Products")
	stateCheckout   = telejoon.NewState[CheckoutData]("Checkout")
	stateSupport    = telejoon.NewState[telejoon.NoData]("Support")
	stateProfile    = telejoon.NewState[telejoon.NoData]("Profile")
	stateAdmin      = telejoon.NewState[telejoon.NoData]("Admin")
	stateBroadcast  = telejoon.NewState[telejoon.NoData]("Broadcast")
	stateOrderReady = telejoon.NewState[telejoon.NoData]("OrderReady")

	// The change-language menu is auto-registered by the engine under this
	// handle (see WithChangeLanguageMenu in main.go); never register your own
	// menu under the same name (engine.Add panics on duplicates).
	stateChooseLanguage = telejoon.NewState[telejoon.NoData]("ChooseLanguage")
)

// CheckoutData is the payload carried into the Checkout state by GoToWith.
type CheckoutData struct {
	ProductID int64
	Qty       int
}

// === Inline menu handles ===

var (
	menuCatalog = telejoon.NewInlineMenuRef("Catalog")
	menuDetails = telejoon.NewInlineMenuRef("ProductDetails")
	menuAbout   = telejoon.NewInlineMenuRef("About")
)

// === Session keys ===
//
// Typed, collision-free keys for per-request storage. Values live for ONE
// update only — the Ctx is rebuilt per update — which is exactly what
// middlewares, conditions, and text builders use to communicate within a
// single request. Anything that must survive across messages (carts, auth)
// belongs in your own storage (see store.go), not in a Key.

var (
	nameKey = telejoon.NewKey[string]("name") // display name, set by middleware

	// Two keys may share a name — identity is unique, they never collide.
	requestTagKey  = telejoon.NewKey[string]("tag")
	requestTagKey2 = telejoon.NewKey[int]("tag")
)

// === Conditions ===
//
// A Cond is memoized: evaluated at most once per request, cached in the Ctx.
// Use If(fn) instead when a predicate is deliberately not shared.

var isAdmin = telejoon.DefineCond(func(ctx *telejoon.Ctx) bool {
	return isAdminUser(ctx.UserID())
})

var isLoggedIn = telejoon.DefineCond(func(ctx *telejoon.Ctx) bool {
	return isLoggedInUser(ctx.UserID())
})
