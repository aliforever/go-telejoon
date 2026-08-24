# Telejoon v3 API

Telejoon v3 is a redesign of the framework around Go 1.27 **generic methods**.
The guiding architecture is *typed rim, dynamic hub*: all generics live at the
registration edge (builders, keys, states, routes) and erase into plain
closures at `engine.Add(...)`, so the dispatch loop stays monomorphic.

Requires **Go 1.27+**.

## The five ideas

1. **Typed states** — states are package variables, not strings. Typos are
   compile errors, and transitions can carry a typed payload.
2. **Typed session storage** — `ctx.Get(key)` / `ctx.Set(key, value)` with
   `Key[T]`, full type inference, no `any` assertions in your code.
3. **Typed part handlers** — one generic method `menu.On(...)` routes by the
   closure's parameter type; no wrapper structs, no runtime type switches.
4. **Typed callback routes** — `menu.Route("del", handler)` returns a typed
   handle; buttons minted with `Do(label, route, args)` encode with the same
   codec the handler decodes with. Producer and consumer cannot drift.
5. **One `Action` result** — handlers return a single `Action`
   (`Next/Stop/Error/Show/Edit/ctx.GoTo/...`), replacing the old
   `(SwitchAction, ShouldPass)` pair.

## Quick start

```go
package main

import (
	"context"
	"fmt"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telejoon"
)

var (
	Welcome  = telejoon.NewState[telejoon.NoData]("Welcome")
	Admin    = telejoon.NewState[telejoon.NoData]("Admin")
	Checkout = telejoon.NewState[CheckoutData]("Checkout")

	Products = telejoon.NewInlineMenuRef("Products")

	IsAdmin = telejoon.DefineCond(func(ctx *telejoon.Ctx) bool {
		return admins[ctx.UserID()]
	})
)

type CheckoutData struct {
	ProductID int64
	Qty       int
}

type BuyArgs struct {
	ProductID int64
}

func main() {
	engine := telejoon.New(telejoon.NewDefaultUserRepository(), Welcome)

	// A typed callback route: the payload type is inferred from the handler.
	products := telejoon.InlineMenuFor(Products, telejoon.S("Products"))

	buy := products.Route("buy", func(ctx *telejoon.Ctx, args BuyArgs) telejoon.Action {
		return ctx.GoToWith(Checkout, CheckoutData{ProductID: args.ProductID, Qty: 1})
	})

	products.ButtonsFunc(func(ctx *telejoon.Ctx) []*telejoon.InlineButton {
		var buttons []*telejoon.InlineButton
		for _, p := range store.List() {
			buttons = append(buttons,
				telejoon.Do(telejoon.S(p.Name), buy, BuyArgs{ProductID: p.ID}).NewRow())
		}
		return buttons
	})

	engine.Add(
		telejoon.Menu(Welcome, telejoon.L("Welcome.Main")).
			Buttons(
				telejoon.ShowInline(telejoon.L("Nav.Products"), Products),
				telejoon.GoTo(telejoon.L("Nav.Admin"), Admin).When(IsAdmin),
			).
			OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
				return ctx.ReplyText("echo: " + text)
			}).
			On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, photo []structs.PhotoSize) telejoon.Action {
				return ctx.ReplyText("nice photo!")
			}),
		telejoon.Menu(Checkout, telejoon.D(func(ctx *telejoon.Ctx) string {
			data, _ := telejoon.StateData(ctx, Checkout)
			return fmt.Sprintf("Buying product %d — send your address:", data.ProductID)
		})).
			OnText(func(ctx *telejoon.Ctx, data *CheckoutData, address string) telejoon.Action {
				store.PlaceOrder(data.ProductID, data.Qty, address)
				return ctx.GoTo(Welcome)
			}),
		products,
	)

	if err := engine.Validate(); err != nil {
		panic(err)
	}

	telejoon.Start(ctx, client, engine)
}
```

## Concepts

### States: `State[D]`

```go
var Welcome  = telejoon.NewState[telejoon.NoData]("Welcome")
var Checkout = telejoon.NewState[CheckoutData]("Checkout")
```

The string name is persisted in your `UserRepository` (unchanged interface —
deployed bots migrate with zero data migration). The type parameter `D`
declares the payload a transition into the state carries:

```go
return ctx.GoTo(Welcome)                          // payload-less states
return ctx.GoToWith(Checkout, CheckoutData{...})  // D inferred from the handle
```

Handlers of a menu receive the resolved payload as `data *D`. Text builders
and conditions (which are state-agnostic) read it via the package function
`telejoon.StateData(ctx, Checkout)`.

Payload persistence: implement `telejoon.StateDataRepository`
(`SetUserStateData`/`GetUserStateData`, JSON-encoded) and register it with
`engine.WithStateDataRepository(repo)`. Without it, payloads are in-process
only and menus receive the zero `D` after a restart.

### Context: `Ctx` and `Key[T]`

`Ctx` carries the client, raw update, current state name, language, and typed
session storage:

```go
var CartKey = telejoon.NewKey[[]CartItem]("cart")

ctx.Set(CartKey, append(cart, item))
cart, ok := ctx.Get(CartKey)       // []CartItem inferred — no annotation
n := ctx.GetOr(CountKey, 0)
ctx.Delete(CartKey)
```

Keys have unique identity: two keys with the same name never collide.

Convenience accessors: `UserID/ChatID/Username/FirstName/Text/CallbackData/
Command()/IsCommand()/IsPhoto().../MessageType()/Language()/SetLanguage()/
Client()`.

Send helpers return `Action` directly: `ctx.ReplyText("...")`,
`ctx.AnswerCallback("...", false)`.

### Actions

```go
telejoon.Next()   // pass to next handler / fall through
telejoon.Stop()   // handled, stop
telejoon.Error(err)
telejoon.Show(menuRef)   // send inline menu
telejoon.Edit(menuRef)   // edit current message into inline menu
ctx.GoTo(state) / ctx.GoToWith(state, data)
ctx.ReplyText(text)
```

### Texts

`Text` is `func(*Ctx) string`; constructors: `S` (static), `L` (localized),
`LP` (localized with params), `D` (deferred), `K` (from a `Key[string]`),
`F` (Sprintf composition).

### Conditions

```go
var IsAdmin = telejoon.DefineCond(func(ctx *telejoon.Ctx) bool {
	return ctx.GetOr(IsAdminKey, false)
})
```

A `Cond` is memoized: evaluated at most once per request, cached in the `Ctx`.
Buttons take `.When(cond)`, `.Unless(cond)`, or `.If(fn)` (non-memoized).

### Buttons

Reply keyboard: `GoTo`, `GoToWith`, `Reply`, `ShowInline`, `Raw`.
Inline keyboard: `Do`, `URL`, `Alert`, `AlertDialog`, `OpenMenu`,
`OpenMenuEdit`, `StateBtn`.

Modifiers (shared by both families): `.When/.Unless/.If/.NewRow()/.Alone()/.Hook(h)`.

Reply-button presses are matched by rendered label (a Telegram limitation),
but the keyboard is rendered once per request and dispatch uses the same
labels; duplicate visible labels are an explicit error, and labels in *all*
configured languages dispatch (so a mid-conversation language switch doesn't
strand the user's keyboard).

### Menus

```go
telejoon.Menu(state, text).
	Buttons(...)/ButtonsFunc(fn).
	OnText(fn).                 // text that matched no button
	On(func(ctx, data, photo []structs.PhotoSize) telejoon.Action { ... }).
	Default(fn).                // fallback
	Use(middleware).            // menu-level middleware
	Formation(2, 1).MaxPerRow(3)

telejoon.InlineMenuFor(ref, text).
	Buttons(...)/ButtonsFunc(fn).
	Route(name, fn, opts...).   // typed route, see below
	Use(middleware).Formation(...).MaxPerRow(...)
```

`engine.Add(menus...)` compiles builders into immutable runtimes
(registration errors panic — they are startup misconfigurations).
`engine.Validate()` checks cross-references (default state, static button
targets).

### Routes and codecs

```go
type DelArgs struct {
	ProductID int64
	Revision  uint16
}

del := products.Route("del", func(ctx *telejoon.Ctx, a DelArgs) telejoon.Action {
	repo.Delete(a.ProductID)
	return telejoon.Edit(Products)
})

button := telejoon.Do(telejoon.S("🗑"), del, DelArgs{ProductID: p.ID, Revision: p.Rev})
```

The default codec encodes structs field-by-field (string, bool, int/uint,
float kinds) compactly. Use `telejoon.WithCodec(telejoon.JSONCodec[DelArgs]())`
for nested payloads, or implement `Codec[A]` yourself. Telegram's 64-byte
`callback_data` limit is enforced when a button is encoded/rendered.
Engine-global routes: `engine.Route("track", fn)` (bound with `Do` as usual).

### Languages

`NewLanguageBuilder` / `NewLanguageConfig` / `WithChangeLanguageMenu` work as
before; the change-language menu is registered automatically:
`engine.WithLanguageConfig(cfg)`.

### Group/channel updates and multiple processors

`NewGroupHandlers()`, `NewChannelHandlers()`, and `NewMultiProcessor(...)` are
unchanged:

```go
multi := telejoon.NewMultiProcessor(engine, telejoon.NewGroupHandlers())
telejoon.Start(ctx, client, multi)
```

## Migrating from v2

| v2 | v3 |
| --- | --- |
| `WithPrivateStateHandlers(repo, "Welcome")` | `telejoon.New(repo, Welcome)` with `var Welcome = telejoon.NewState[telejoon.NoData]("Welcome")` |
| `AddStaticMenu("X", menu)` | `telejoon.Menu(stateX, text)...` + `engine.Add(...)` (menu knows its state) |
| `AddInlineMenu("X", menu)` | `telejoon.InlineMenuFor(refX, text)...` + `engine.Add(...)` |
| `(SwitchAction, ShouldPass)` returns | single `Action`: `Next()/Stop()/Error()/...` |
| `NewSwitchActionState("X")` | `ctx.GoTo(stateX)` / `ctx.GoToWith(stateX, data)` |
| `NewSwitchActionInlineMenu("X", edit)` | `telejoon.Show(refX)` / `telejoon.Edit(refX)` |
| `update.Set("k", v)` / `GetString(...)` | `ctx.Set(K, v)` / `ctx.Get(K)` with `var K = telejoon.NewKey[T]("k")` |
| `NewStaticText/NewLanguageKeyText/NewDeferredText/NewUpdateKeyText/NewTextBuilderF` | `S/L/D/K/F` |
| `buttons.GoTo(...).WhenDefined("isAdmin")` | `telejoon.GoTo(...).When(IsAdmin)` with `var IsAdmin = telejoon.DefineCond(fn)` |
| `TextButton/StateButton/InlineMenuButton/RawButton` + Conditional/DefinedConditional/VsDefinedConditional variants | `Reply/GoTo/ShowInline/Raw` + `.When/.Unless/.If` modifiers |
| `NewDynamicHandlerPhoto(fn)` etc. (13 wrappers) | `.On(func(ctx, data, photo []structs.PhotoSize) ...)` — type inferred |
| `NewStaticMenu(text, builder, ...Handler)` grab-bag | named builder methods (`.Use`, `.OnText`, `.On`, `.Default`) |
| `inline.Callback(text.S("Del"), handler).Data("del:1")` | `r := menu.Route("del", handler)` + `telejoon.Do(label, r, args)` |
| `AddCallbackQueryHandler("data", fn)` | `engine.Route("data", fn)` |
| `NewMiddleware(fn)` | pass the `Handler` func directly to `.Use(...)` |
| `AddTextCommand/AddStateCommand/...` | dropped (was never dispatched) — use `OnText` + `ctx.Command()` |
| `NewButtonOptions(breakBefore, breakAfter)` | `.NewRow()` / `.Alone()` |
| `SetButtonFormation/SetMaxButtonPerRow` | `.Formation(...)` / `.MaxPerRow(...)` on the menu builder |
| `panic("invalid Handler type")` at registration | compile error |
| `no_handler_for_state` at runtime | `engine.Validate()` at startup |
