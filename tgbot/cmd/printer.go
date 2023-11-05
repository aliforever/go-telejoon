package cmd

import "fmt"

type Printer struct {
}

// NewPrinter returns a new Printer.
func NewPrinter() *Printer {
	return &Printer{}
}

// PrintDeferredTextFunction is a function type that is used to print deferred text function.
func (p *Printer) PrintDeferredTextFunction() {
	fn := `func (b *Bot) DeferredText(update *telejoon.StateUpdate) string {
	return "deferred text"
}`
	fmt.Println(fn)
}

func (p *Printer) PrintDeferredActionHandlerFunction() {
	fn := `func (b *Bot) DeferredActionHandler(
	client *tgbotapi.TelegramBot,
	update *telejoon.StateUpdate,
) (telejoon.SwitchAction, bool) {
	return telejoon.NewSwitchActionState("Welcome"), true
}`

	fmt.Println(fn)
}
