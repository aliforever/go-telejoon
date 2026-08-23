# shopbot — telejoon v3 example

A small shop bot demonstrating the whole telejoon v3 API. Run it from the
repository root:

```bash
BOT_TOKEN=... go run ./examples/shopbot
# optional: ERROR_GROUP_ID=<group id> for error reports
```

Files: `states.go` (typed state handles, inline refs, session keys, conds),
`store.go` (demo data + the StateDataRepository), `menus.go` (state menus),
`inline.go` (inline menus + typed routes), `main.go` (wiring).

## The tour

1. **First contact**: force-choose language (auto-registered menu), then land
   on the localized Welcome menu.
2. **Welcome**: `GoTo`/`Reply`/`Raw` buttons, memoized `isAdmin` cond, `/whoami`,
   `/order 42`, `/start buy_2` deep link, `/boom` panic demo (admin),
   `/cancel` global middleware.
3. **Products**: `ButtonsFunc` dynamic keyboard, featured-product `GoToWith`
   with a stock-check `Hook`, cart button via `If`, catalog via `ShowInline`.
4. **Catalog (inline)**: one typed `buy` route button per product (positional
   codec), `details` route (JSON codec), `refresh` (NoData payload), toast
   answer before `GoToWith`.
5. **Checkout**: payload renders the prompt (`StateData` in a `D` text),
   `➕ Quantity` re-enters the state with a new payload, any text is the
   address → order placed → back to Welcome. The payload survives to that
   next message because of `WithStateDataRepository`.
6. **Details (inline)**: toast vs popup (`Alert`/`AlertDialog`), `URL`,
   global `track` route, custom base36 codec route, `OpenMenu` (new message)
   vs `OpenMenuEdit` (in place), `StateBtn` back to the reply-keyboard world.
7. **Support**: typed part handlers for photo/document/contact/dice + `Default`.
8. **Profile**: login/logout toggles with `When`/`Unless` on `Raw` buttons,
   matched by rendered (localized) labels; `F`+`K`+`D` text composition.
9. **Admin** (user ID 1): menu-middleware gate with redirect, broadcast flow,
   `/orderready` (external `SwitchUserState` push), `/sendcatalog`
   (external `SendInlineMenu`).
10. **Language**: switch to فارسی, then press a button on the OLD keyboard —
    labels in all configured languages still dispatch. RTL reverses row order.
11. **Groups/channels**: add the bot to a group (`/shop` keyword reply) and a
    channel (post logging) to see `MultiProcessor` routing.

## Things to notice in the code

- Session `Key[T]` storage is **per-request**, not a user session — the cart
  lives in a real store (`store.go`), the `K` text demo reads a value set by
  middleware in the same update.
- Custom `Do` route handlers must answer the callback themselves
  (`ctx.AnswerCallback`); built-in alert/menu/state buttons auto-answer.
- `engine.Validate()` catches broken static references at startup; dynamic
  `ButtonsFunc` output is validated at render (e.g. duplicate labels).
