package telejoon

// Response helpers for cleaner handler return values.
// These replace verbose return statements like `return nil, true`.

// Continue returns a response that continues to the next handler.
// Use when the current handler doesn't want to stop processing.
//
// Example:
//
//	func(c *tgbotapi.TelegramBot, u *StateUpdate) (SwitchAction, ShouldPass) {
//	    log.Println("Middleware executed")
//	    return Continue()
//	}
func Continue() (SwitchAction, ShouldPass) {
	return nil, true
}

// Stop returns a response that stops processing.
// Use when the current handler has fully handled the update.
//
// Example:
//
//	func(c *tgbotapi.TelegramBot, u *StateUpdate) (SwitchAction, ShouldPass) {
//	    c.Send(c.Message().SetChatId(u.Update.From().Id).SetText("Handled!"))
//	    return Stop()
//	}
func Stop() (SwitchAction, ShouldPass) {
	return nil, false
}

// GoTo switches to a state and stops processing.
//
// Example:
//
//	func(c *tgbotapi.TelegramBot, u *StateUpdate) (SwitchAction, ShouldPass) {
//	    if u.Update.Message.Text == "admin" {
//	        return GoTo("AdminPanel")
//	    }
//	    return Continue()
//	}
func GoTo(state string) (SwitchAction, ShouldPass) {
	return NewSwitchActionState(state), false
}

// ShowInline shows an inline menu as a new message.
//
// Example:
//
//	return ShowInline("SettingsMenu")
func ShowInline(menu string) (SwitchAction, ShouldPass) {
	return NewSwitchActionInlineMenu(menu, false), false
}

// EditInline edits the current message to an inline menu.
//
// Example:
//
//	return EditInline("SettingsMenu")
func EditInline(menu string) (SwitchAction, ShouldPass) {
	return NewSwitchActionInlineMenu(menu, true), false
}
