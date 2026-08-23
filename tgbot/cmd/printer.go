package cmd

import "fmt"

type Printer struct {
}

// NewPrinter returns a new Printer.
func NewPrinter() *Printer {
	return &Printer{}
}

// PrintDeferredTextFunction prints a sample deferred text function.
func (p *Printer) PrintDeferredTextFunction() {
	fn := `func (b *Bot) DeferredText(ctx *telejoon.Ctx) string {
	return "deferred text"
}`
	fmt.Println(fn)
}

// PrintDeferredActionHandlerFunction prints a sample menu text handler.
func (p *Printer) PrintDeferredActionHandlerFunction() {
	fn := `func (b *Bot) WelcomeTextHandler(ctx *telejoon.Ctx, data *telejoon.NoData, text string) telejoon.Action {
	return ctx.GoTo(stateWelcome)
}`

	fmt.Println(fn)
}
