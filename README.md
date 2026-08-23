# go-telejoon

Telegram bot framework built around Go 1.27 **generic methods**: typed states
with payloads, typed session storage, typed callback routes, and type-inferred
message handlers — with a small monomorphic engine at the core.

Requires **Go 1.27+**. See [API_V3.md](API_V3.md) for the full guide and the
v2 → v3 migration table.

```go
var (
	Welcome  = telejoon.NewState[telejoon.NoData]("Welcome")
	Checkout = telejoon.NewState[CheckoutData]("Checkout")
)

engine := telejoon.New(userRepo, Welcome)

buy := products.Route("buy", func(ctx *telejoon.Ctx, args BuyArgs) telejoon.Action {
	return ctx.GoToWith(Checkout, CheckoutData{ProductID: args.ProductID, Qty: 1})
})

engine.Add(
	telejoon.Menu(Welcome, telejoon.L("Welcome.Main")).
		Buttons(telejoon.GoTo(telejoon.L("Nav.Admin"), Admin).When(IsAdmin)).
		On(func(ctx *telejoon.Ctx, _ *telejoon.NoData, photo []structs.PhotoSize) telejoon.Action {
			return ctx.ReplyText("nice photo!")
		}),
	products,
)

telejoon.Start(ctx, client, engine)
```

## Note

- This is a work in progress and is not ready for production use.

## Docs

[Here](https://pkg.go.dev/github.com/aliforever/go-telejoon)

## Install Project Generator

```bash
go install github.com/aliforever/go-telejoon/tgbot@latest
```

## Generate Project

```bash
tgbot --token=BOT_TOKEN_HERE --module_path=MODULE_PATH_HERE
```
