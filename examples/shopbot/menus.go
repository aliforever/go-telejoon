package main

import (
	"fmt"
	"strings"

	"github.com/aliforever/go-telegram-bot-api/structs"
	"github.com/aliforever/go-telejoon"
)

// === Typed message handles ===
//
// NewMsg/NewMsgP declare localized messages as package variables instead of
// scattering key strings: one place to rename a key, and engine.Validate
// checks declared NewMsg keys against the locales at startup. MsgP's type
// parameter is the params struct — its JSON field names are the template
// variable names, so a template and its data cannot drift apart.

var welcomeMain = telejoon.NewMsgP[welcomeParams]("Welcome.Main")

type welcomeParams struct {
	Name string `json:"Name"` // {{.Name}} in the locale files
}

var checkoutPrompt = telejoon.NewMsgP[checkoutParams]("Checkout.Text")

type checkoutParams struct {
	ProductID int64 `json:"ProductID"`
	Qty       int   `json:"Qty"`
}

// === Welcome ===
//
// The home menu: static buttons of every kind, a memoized condition, a
// per-request middleware, and command parsing in OnText.

func welcomeMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateWelcome, telejoon.D(welcomeText)).
		Use(func(ctx *telejoon.Ctx) telejoon.Action {
			// Menu-level middleware: runs after engine middlewares, before
			// any dispatch. Session values set here are visible to this
			// request's texts and handlers — but not to the next update.
			ctx.Set(nameKey, ctx.FirstName())

			return telejoon.Next()
		}).
		Buttons(
			telejoon.GoTo(telejoon.L("Welcome.Products"), stateProducts),
			telejoon.GoTo(telejoon.L("Welcome.Support"), stateSupport),
			telejoon.GoTo(telejoon.L("Welcome.Profile"), stateProfile),
			// Reply sends a text without leaving the state.
			telejoon.Reply(telejoon.L("Welcome.Help"), telejoon.L("Welcome.HelpBody")),
			// The change-language menu is auto-registered by the engine;
			// address it by a handle with the same string name.
			telejoon.GoTo(telejoon.S("🌐 Language"), stateChooseLanguage),
			// When: the memoized isAdmin cond is evaluated once per keyboard
			// render, no matter how many buttons use it. Alone: own row.
			telejoon.GoTo(telejoon.L("Welcome.Admin"), stateAdmin).When(isAdmin).Alone(),
			// Raw renders like any button but has no action: the press falls
			// through to OnText (matched by its rendered label below).
			telejoon.Raw(telejoon.L("General.Crash")).When(isAdmin),
		).
		// Formation(2, 2): rows of 2, then the last entry repeats; MaxPerRow
		// applies when the formation is exhausted... see chunkIntoRows docs.
		Formation(2, 2).
		MaxPerRow(3).
		OnText(welcomeOnText)
}

func welcomeText(ctx *telejoon.Ctx) string {
	name := ctx.GetOr(nameKey, "")
	if name == "" {
		name = ctx.FirstName()
	}

	// Typed params: missing translations fall back to the default language;
	// a key missing everywhere renders as the key itself — never a panic.
	return welcomeMain.T(welcomeParams{Name: name})(ctx)
}

func welcomeOnText(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
	// Commands are ordinary text messages in v3 — parse them yourself.
	if ctx.IsCommand() {
		switch cmd := ctx.Command(); cmd.Name {
		case "start":
			// Deep link: /start buy_2 jumps straight into checkout.
			if id, ok := strings.CutPrefix(cmd.ArgOr(0, ""), "buy_"); ok {
				if p := productByID(parseID(id)); p != nil {
					return ctx.GoToWith(stateCheckout, CheckoutData{ProductID: p.ID, Qty: 1})
				}
			}

			// Next falls through to the engine's render step: the menu text
			// and keyboard are (re)sent. A "show the menu again" trick.
			return telejoon.Next()
		case "whoami":
			return ctx.ReplyText(fmt.Sprintf(
				"user=%d chat=%d username=@%s name=%s",
				ctx.UserID(), ctx.ChatID(), ctx.Username(), ctx.FirstName()))
		case "order":
			// Arg/ArgOr/RawArgs/ArgCount: see ParseCommand.
			return ctx.ReplyText(fmt.Sprintf("order %s: shipped (demo). raw args: %q",
				cmd.ArgOr(0, "?"), cmd.RawArgs))
		case "boom":
			// Panic demo: recovered per update by the engine's panic handler;
			// the bot stays up. (In a real bot, hide this behind admin checks.)
			panic("boom — intentional demo panic")
		}
	}

	// Raw button press: compare against the RENDERED label — hardcoding the
	// English string would break for users on another language. The button is
	// admin-only (When(isAdmin)), and the typed text must be gated the same
	// way: callback_data and message texts are client-controlled, so button
	// visibility alone is never an authorization check.
	if isAdminUser(ctx.UserID()) && text == telejoon.L("General.Crash")(ctx) {
		panic("boom — intentional demo panic via raw button")
	}

	return ctx.ReplyText("echo: " + text)
}

// === Products ===
//
// Dynamic reply keyboard (ButtonsFunc replaces Buttons entirely — return the
// navigation buttons from the function too), a Hook gating a GoToWith, and
// the If modifier for a request-specific predicate.

func productsMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateProducts,
		// F composes Texts with fmt.Sprintf; D defers per request.
		telejoon.F("%s\n%s", telejoon.L("Products.Title"), telejoon.D(cartSummary)),
	).
		ButtonsFunc(func(ctx *telejoon.Ctx, _ *telejoon.NoData) []*telejoon.Button {
			buttons := []*telejoon.Button{
				telejoon.ShowInline(telejoon.S("🛍 Open catalog"), menuCatalog),
				// A GoToWith reply button with the payload baked in. The Hook
				// runs after the press, before the transition: returning
				// anything but Next cancels the button's action.
				telejoon.GoToWith(telejoon.S("⭐ Featured: Golden Mug"), stateCheckout,
					CheckoutData{ProductID: 3, Qty: 1}).
					Hook(func(ctx *telejoon.Ctx) telejoon.Action {
						if !productByID(3).InStock {
							return ctx.ReplyText(telejoon.L("Products.OutOfStock")(ctx))
						}

						return telejoon.Next()
					}),
			}

			// If: a non-memoized, request-specific predicate.
			if cart := cartOf(ctx.UserID()); len(cart) > 0 {
				buttons = append(buttons,
					telejoon.Raw(telejoon.S(fmt.Sprintf("🧺 Cart (%d)", len(cart)))).
						If(func(ctx *telejoon.Ctx) bool { return len(cartOf(ctx.UserID())) > 0 }))
			}

			return append(buttons, telejoon.GoTo(telejoon.L("General.Back"), stateWelcome))
		}).
		OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
			if strings.HasPrefix(text, "🧺 Cart") {
				var lines []string
				for _, id := range cartOf(ctx.UserID()) {
					if p := productByID(id); p != nil {
						lines = append(lines, fmt.Sprintf("- %s ($%.2f)", p.Name, p.Price))
					}
				}

				return ctx.ReplyText(strings.Join(lines, "\n"))
			}

			return telejoon.Next() // unknown text: re-render the menu
		})
}

func cartSummary(ctx *telejoon.Ctx) string {
	return fmt.Sprintf("cart: %d item(s)", len(cartOf(ctx.UserID())))
}

// === Checkout ===
//
// The typed-payload showcase: the payload renders the prompt (via StateData,
// since texts are state-agnostic), drives the keyboard (via data *D), and is
// still there when the address arrives on the NEXT update — thanks to
// WithStateDataRepository (see store.go for why that is required).

func checkoutMenu() *telejoon.MenuBuilder[CheckoutData] {
	return telejoon.Menu(stateCheckout, telejoon.D(checkoutText)).
		ButtonsFunc(func(ctx *telejoon.Ctx, data *CheckoutData) []*telejoon.Button {
			return []*telejoon.Button{
				telejoon.Raw(telejoon.S(fmt.Sprintf("➕ Quantity: %d", data.Qty))),
				telejoon.Raw(telejoon.S("❌ Cancel")).Alone(),
				telejoon.GoTo(telejoon.L("General.Back"), stateProducts),
			}
		}).
		OnText(func(ctx *telejoon.Ctx, data *CheckoutData, text string) telejoon.Action {
			switch {
			case strings.HasPrefix(text, "➕"):
				// Re-enter the same state with a new payload.
				return ctx.GoToWith(stateCheckout, CheckoutData{ProductID: data.ProductID, Qty: data.Qty + 1})
			case text == "❌ Cancel":
				return ctx.GoTo(stateProducts)
			}

			// Any other text is the delivery address: place the order.
			placeOrder(ctx.UserID(), order{ProductID: data.ProductID, Qty: data.Qty, Address: text})

			// ctx.Client() is the escape hatch for operations that no Action
			// covers; here we send a confirmation AND transition afterwards.
			done := telejoon.L("Checkout.Done")(ctx)
			if _, err := ctx.Client().Message().
				ChatID(ctx.ChatID()).
				Text(done).
				Send(ctx.Context()); err != nil {
				return telejoon.Error(err)
			}

			return ctx.GoTo(stateWelcome)
		}).
		// Typed part handlers: P is inferred from the closure parameter.
		On(func(ctx *telejoon.Ctx, data *CheckoutData, loc *structs.Location) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Support.GotLocation")(ctx),
				loc.Latitude, loc.Longitude))
		}).
		Default(func(ctx *telejoon.Ctx, data *CheckoutData) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf("can't handle %s here — send the address as text",
				ctx.MessageType()))
		})
}

func checkoutText(ctx *telejoon.Ctx) string {
	// Texts and conditions are state-agnostic, so they read the payload with
	// the package function. ok is false when ctx.State is another state.
	data, ok := telejoon.StateData(ctx, stateCheckout)
	if !ok {
		return "checkout"
	}

	return checkoutPrompt.T(checkoutParams{ProductID: data.ProductID, Qty: data.Qty})(ctx)
}

// === Support ===
//
// Media intake: one On handler per message part type. The full Part union is
// photo, video, document, voice, audio, sticker, location, contact,
// video note, venue, poll, and dice — these five are representative.

func supportMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateSupport, telejoon.L("Support.Text")).
		Buttons(telejoon.GoTo(telejoon.L("General.Back"), stateWelcome)).
		OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Support.GotText")(ctx), text))
		}).
		On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, photo []structs.PhotoSize) telejoon.Action {
			// Note: photos carry their caption in ctx.Update.Message.Caption,
			// not in ctx.Text() — OnText never fires for them.
			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Support.GotPhoto")(ctx), len(photo)))
		}).
		On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, doc *structs.Document) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Support.GotDocument")(ctx), doc.FileName))
		}).
		On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, contact *structs.Contact) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Support.GotContact")(ctx), contact.FirstName))
		}).
		On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, loc *structs.Location) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Support.GotLocation")(ctx),
				loc.Latitude, loc.Longitude))
		}).
		On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, dice *structs.Dice) telejoon.Action {
			return ctx.ReplyText(fmt.Sprintf("you rolled %d", dice.Value))
		}).
		Default(func(ctx *telejoon.Ctx, _ *telejoon.NoData) telejoon.Action {
			return ctx.ReplyText(telejoon.L("Support.GotOther")(ctx) + " (" + ctx.MessageType() + ")")
		})
}

// === Profile ===
//
// Login/logout toggling with When/Unless on Raw buttons, matched in OnText
// by their rendered (localized) labels.

func profileMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateProfile, telejoon.F("%s\n%s\n(session name via K: %q)",
		telejoon.L("Profile.Text"),
		telejoon.D(func(ctx *telejoon.Ctx) string {
			if isLoggedInUser(ctx.UserID()) {
				// Accounts may have no username; fall back to the first name.
				name := ctx.Username()
				if name == "" {
					name = ctx.FirstName()
				}

				return fmt.Sprintf(telejoon.L("Profile.LoggedIn")(ctx), name)
			}

			return telejoon.L("Profile.LoggedOut")(ctx)
		}),
		// K renders a typed session value — set by the engine middleware
		// earlier in THIS request (session storage is per-update).
		telejoon.K(nameKey),
	)).
		Buttons(
			telejoon.Raw(telejoon.L("General.Login")).Unless(isLoggedIn),
			telejoon.Raw(telejoon.L("General.Logout")).When(isLoggedIn),
			telejoon.GoTo(telejoon.L("General.Back"), stateWelcome),
		).
		OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
			switch text {
			case telejoon.L("General.Login")(ctx):
				loggedInUsers.Store(ctx.UserID(), true)
			case telejoon.L("General.Logout")(ctx):
				loggedInUsers.Delete(ctx.UserID())
			default:
				return telejoon.Next()
			}

			// Re-enter the state to re-render with the new visibility.
			return ctx.GoTo(stateProfile)
		})
}

// === Admin ===
//
// A menu-level middleware gate (redirecting non-admins), plus the external
// engine operations SwitchUserState / SendInlineMenu driven by commands.

func adminMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateAdmin, telejoon.L("Admin.Text")).
		Use(func(ctx *telejoon.Ctx) telejoon.Action {
			if !isAdminUser(ctx.UserID()) {
				// Middleware redirect. Two menus redirecting to each other
				// would die loudly after 8 hops (maxStateSwitchDepth), not
				// recurse forever.
				return ctx.GoTo(stateWelcome)
			}

			return telejoon.Next()
		}).
		Buttons(
			telejoon.GoTo(telejoon.L("Admin.Broadcast"), stateBroadcast),
			telejoon.GoTo(telejoon.L("General.Back"), stateWelcome),
		).
		OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
			if !ctx.IsCommand() {
				return telejoon.Next()
			}

			switch ctx.Command().Name {
			case "orderready":
				// External transition with no update in flight for the target
				// user (here: self, for demo). Only State[NoData] targets —
				// typed payloads belong to in-processing GoToWith.
				if err := botEngine.SwitchUserState(ctx.Context(), ctx.Client(), ctx.UserID(), stateOrderReady); err != nil {
					return telejoon.Error(err)
				}

				return telejoon.Stop()
			case "sendcatalog":
				// External inline-menu push for a given update.
				if err := botEngine.SendInlineMenu(ctx.Context(), ctx.Client(), ctx.Update, menuCatalog, false); err != nil {
					return telejoon.Error(err)
				}

				return telejoon.Stop()
			}

			return telejoon.Next()
		})
}

// === Broadcast ===

func broadcastMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateBroadcast, telejoon.L("Admin.BroadcastPrompt")).
		Buttons(telejoon.GoTo(telejoon.L("General.Back"), stateAdmin)).
		OnText(func(ctx *telejoon.Ctx, _ *telejoon.NoData, text string) telejoon.Action {
			count := 0
			loggedInUsers.Range(func(_, _ any) bool {
				count++

				return true
			})

			// Demo only: nothing is actually sent.
			_ = text

			return ctx.ReplyText(fmt.Sprintf(telejoon.L("Admin.BroadcastDone")(ctx), count))
		})
}

// === OrderReady ===
//
// Target of the external SwitchUserState push.

func orderReadyMenu() *telejoon.MenuBuilder[telejoon.NoData] {
	return telejoon.Menu(stateOrderReady, telejoon.S("🎉 Your order is ready for pickup!")).
		Buttons(telejoon.GoTo(telejoon.L("General.Back"), stateWelcome))
}

func parseID(raw string) int64 {
	var id int64

	_, _ = fmt.Sscanf(raw, "%d", &id)

	return id
}
