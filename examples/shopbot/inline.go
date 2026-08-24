package main

import (
	"fmt"

	"github.com/aliforever/go-telejoon"
)

// === Route payload types ===

// buyArgs uses the default positional codec: a flat struct of primitive
// fields, encoded compactly. (Slices or nested structs would fail to encode —
// use JSONCodec for those.)
type buyArgs struct {
	ProductID int64
}

// detailsArgs uses JSONCodec. JSON is verbose, and the 64-byte
// callback_data limit covers menu name + route name + escaping + payload —
// {"i":1,"s":"c"} percent-escapes to ~37 bytes, so "Catalog:details:<payload>"
// just fits. With the original long field names and "catalog" as the source
// it exceeded the limit and the button failed to render. Size accordingly,
// or use the positional codec/custom Codec for anything bigger.
type detailsArgs struct {
	ProductID int64  `json:"i"`
	Source    string `json:"s"`
}

// trackRoute is the engine-global "track" route, registered in main. Buttons
// minted from it work from any inline menu, but note: a global route is NOT
// owned by the menu carrying the button — menu middleware does not run for
// it, and its name may not collide with an inline-menu name.
var trackRoute telejoon.Route[int64]

// === Catalog (dynamic inline keyboard) ===

func catalogMenu() *telejoon.InlineMenuBuilder {
	catalog := telejoon.InlineMenuFor(menuCatalog, telejoon.L("Products.Title")).
		Use(func(ctx *telejoon.Ctx) telejoon.Action {
			// Inline-menu middleware runs when the menu's callbacks fire and
			// when the menu renders — but at most once per update, even for
			// the refresh idiom below (route -> Edit of the same menu).
			return telejoon.Next()
		})

	// Menu-scoped typed routes. The payload type is inferred from the
	// handler; buttons minted with Do encode with the same codec the handler
	// decodes with, so the two can never drift.
	buy := catalog.Route("buy", func(ctx *telejoon.Ctx, args buyArgs) telejoon.Action {
		p := productByID(args.ProductID)
		if p == nil {
			// Error routes to Options.Logger / ErrorGroupID.
			return telejoon.Error(fmt.Errorf("buy: unknown product %d", args.ProductID))
		}

		if !p.InStock {
			// AnswerCallback with showAlert=true renders as a popup dialog.
			return ctx.AnswerCallback(telejoon.L("Products.OutOfStock")(ctx), true)
		}

		addToCart(ctx.UserID(), p.ID)

		// Answering with a text shows a toast; a plain no-op answer is sent
		// automatically after the route when the handler did not answer, so
		// the client's spinner never runs until timeout.
		_ = ctx.AnswerCallback(telejoon.L("Products.Added")(ctx), false)

		return ctx.GoToWith(stateCheckout, CheckoutData{ProductID: p.ID, Qty: 1})
	})

	details := catalog.Route("details", func(ctx *telejoon.Ctx, args detailsArgs) telejoon.Action {
		p := productByID(args.ProductID)
		if p == nil {
			return telejoon.Error(fmt.Errorf("details: unknown product %d", args.ProductID))
		}

		return ctx.AnswerCallback(fmt.Sprintf("%s — $%.2f (in stock: %v)",
			p.Name, p.Price, p.InStock), true)
	}, telejoon.WithCodec(telejoon.JSONCodec[detailsArgs]()))

	// A NoData payload route — the "refresh" idiom. The callback is answered
	// automatically, and the menu's middleware chain does not re-run for the
	// re-render: a menu's chain runs at most once per update.
	refresh := catalog.Route("refresh", func(ctx *telejoon.Ctx, _ telejoon.NoData) telejoon.Action {
		return telejoon.Edit(menuCatalog) // re-render in place
	})

	catalog.ButtonsFunc(func(ctx *telejoon.Ctx) []*telejoon.InlineButton {
		var buttons []*telejoon.InlineButton

		for _, p := range products {
			buttons = append(buttons,
				telejoon.Do(telejoon.S(fmt.Sprintf("%s — $%.2f", p.Name, p.Price)),
					buy, buyArgs{ProductID: p.ID}).NewRow())
		}

		return append(buttons,
			telejoon.Do(telejoon.L("General.Details"), details,
				detailsArgs{ProductID: 1, Source: "c"}),
			telejoon.Do(telejoon.L("General.Refresh"), refresh, telejoon.NoData{}),
			telejoon.OpenMenuEdit(telejoon.S("🔎 More…"), menuDetails),
		)
	})

	return catalog
}

// === ProductDetails (navigation chain + built-in inline buttons) ===

func detailsMenu() *telejoon.InlineMenuBuilder {
	menu := telejoon.InlineMenuFor(menuDetails, telejoon.S(
		"Details & navigation demos:"))

	// A route with a custom codec (base36 order IDs — shorter payloads).
	reorder := menu.Route("reorder", func(ctx *telejoon.Ctx, orderID int64) telejoon.Action {
		return ctx.AnswerCallback(fmt.Sprintf("order #%d reordered (base36 codec)", orderID), false)
	}, telejoon.WithCodec(base36Codec{}))

	menu.Buttons(
		// Alert: toast notification. Keep the text SHORT and ASCII — it is
		// embedded (percent-escaped) in callback_data and counts against the
		// 64-byte limit. For long or localized texts, use a Do route +
		// ctx.AnswerCallback instead.
		telejoon.Alert(telejoon.S("ℹ️ Stock (toast)"), "3 in stock"),
		// AlertDialog: same slot, but a popup dialog.
		telejoon.AlertDialog(telejoon.S("⚠️ Note (popup)"), "No refunds on demo mugs!"),
		// URL buttons open a link; they carry no callback_data at all.
		telejoon.URL(telejoon.S("🌐 Terms"), telejoon.S("https://example.com/terms")),
		// A button bound to the engine-global track route.
		telejoon.Do(telejoon.S("🚚 Track order #1"), trackRoute, int64(1)),
		// Custom-codec route button.
		telejoon.Do(telejoon.S("🔁 Reorder #7"), reorder, int64(7)),
		// OpenMenu sends the target as a NEW message...
		telejoon.OpenMenu(telejoon.S("📄 About (new message)"), menuAbout),
		// ...while OpenMenuEdit edits the current one — usually what you want
		// for drill-down navigation.
		telejoon.OpenMenuEdit(telejoon.S("⬅️ Catalog"), menuCatalog),
		// StateBtn leaves the inline world and switches the user's state.
		telejoon.StateBtn(telejoon.S("🆘 Contact support"), stateSupport),
	).MaxPerRow(2)

	return menu
}

// === About (target of the OpenMenu contrast) ===

func aboutMenu() *telejoon.InlineMenuBuilder {
	return telejoon.InlineMenuFor(menuAbout, telejoon.S(
		"This message was sent as a NEW message (OpenMenu), not an edit.")).
		Buttons(telejoon.OpenMenuEdit(telejoon.S("⬅️ Back to details"), menuDetails))
}
