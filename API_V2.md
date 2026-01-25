# Telejoon v2 API Documentation

A comprehensive guide to the new fluent APIs for buttons, menus, and handlers.

---

## Quick Start

```go
import (
    "github.com/aliforever/go-telejoon"
    "github.com/aliforever/go-telejoon/buttons"
    "github.com/aliforever/go-telejoon/inline"
    "github.com/aliforever/go-telejoon/menu"
    "github.com/aliforever/go-telejoon/text"
)
```

---

## 1. Text Package

Type-safe text builders for button labels.

```go
text.S("static text")     // Static text
text.L("Lang.Key")        // Language key (localized)
text.D(func(u) string)    // Deferred (computed at render)
text.K("contextKey")      // From StateUpdate storage
```

---

## 2. Buttons Package

### Creating Reply Keyboard Buttons

```go
buttons.GoTo(label, state)      // Switch to state
buttons.Reply(label, response)  // Send text response
buttons.Show(label, menu)       // Open inline menu
buttons.Raw(label)              // No action (for dynamic handlers)
```

### Modifiers

```go
// Conditions
.If(bool)              // Compile-time include/exclude
.When(func)            // Render-time visibility
.WhenDefined(name)     // Named condition from builder
.Unless(func)          // Inverse condition
.UnlessDefined(name)   // Inverse named condition

// Layout
.NewRow()              // Start on new row
.Alone()               // Own row (break before + after)

// Hooks
.Before(hook)          // Pre-action handler
```

### Builder

```go
buttons.Build(
    buttons.GoTo(text.L("Nav.Home"), "Welcome"),
    buttons.GoTo(text.S("Admin"), "Admin").WhenDefined("isAdmin"),
    buttons.GoTo(text.S("Back"), "Main").Alone(),
).
    Define("isAdmin", isAdminFunc).
    Formation(2, 1).
    MaxPerRow(3).
    Build()
```

---

## 3. Inline Package

### Creating Inline Keyboard Buttons

```go
inline.URL(label, url)                  // Open URL
inline.Alert(label, text)               // Toast notification
inline.Confirm(label, text)             // Popup dialog
inline.Menu(label, menuName)            // Navigate (new message)
inline.MenuEdit(label, menuName)        // Navigate (edit message)
inline.State(label, stateName)          // Switch state
inline.Callback(label, handler)         // Custom handler
```

### Modifiers

```go
.Data(string)          // Set callback data
.DataD(func)           // Dynamic callback data
.If(bool)              // Compile-time condition
.When(func)            // Render-time condition
.NewRow()              // Start on new row
.Alone()               // Own row
```

### Builder

```go
inline.Build(
    inline.URL(text.S("Website"), "https://example.com"),
    inline.Callback(text.S("Delete"), handler).Data("del:123"),
).MaxPerRow(2).Build()
```

---

## 4. Menu Package

### Static Menu (Reply Keyboard)

```go
menu.Static(text.L("Welcome.Message")).
    Buttons(
        buttons.GoTo(text.L("Nav.Home"), "Welcome"),
        buttons.GoTo(text.S("Settings"), "Settings"),
    ).
    Middleware(logMiddleware).
    OnText(handleText).
    OnPhoto(handlePhoto).
    OnDocument(handleDocument).
    OnAny(handleDefault).
    Build()
```

### Inline Menu

```go
menu.Inline(text.S("Choose option")).
    Buttons(
        inline.URL(text.S("Website"), "https://..."),
        inline.Callback(text.S("Action"), handler),
    ).
    Middleware(checkAuth).
    Build()
```

---

## 5. Handler Response Helpers

Clean return statements for handlers.

```go
// Instead of: return nil, true
return telejoon.Continue()

// Instead of: return nil, false
return telejoon.Stop()

// Instead of: return NewSwitchActionState("Admin"), false
return telejoon.GoTo("Admin")

// Switch to inline menu
return telejoon.ShowInline("Menu")   // New message
return telejoon.EditInline("Menu")   // Edit current message
```

---

## 6. StateUpdate Improvements

### Typed Getters

```go
u.GetString("key", "default")
u.GetInt("count", 0)
u.GetInt64("id", 0)
u.GetBool("active", false)
u.Has("key")
u.Delete("key")
u.Clear()
```

### Convenience Accessors

```go
u.UserID()       // Instead of: u.Update.From().Id
u.ChatID()       // Instead of: u.Update.Chat().Id
u.Username()
u.FirstName()
u.Text()         // Instead of: u.Update.Message.Text
u.CallbackData()
u.CallbackID()
```

### Message Type Detection

```go
u.IsText()
u.IsPhoto()
u.IsDocument()
u.IsVideo()
u.IsVoice()
u.IsAudio()
u.IsLocation()
u.IsContact()
u.IsSticker()
u.IsCallback()
u.MessageType()  // Returns: "text", "photo", etc.
```

---

## Complete Example

```go
// Define a welcome menu
welcomeMenu := menu.Static(text.L("Welcome.Message")).
    Buttons(
        buttons.GoTo(text.L("Nav.Products"), "Products"),
        buttons.GoTo(text.L("Nav.Settings"), "Settings"),
        buttons.GoTo(text.S("Admin"), "Admin").WhenDefined("isAdmin"),
        buttons.Reply(text.S("Help"), "Here is help text..."),
    ).
    Middleware(func(c *tgbotapi.TelegramBot, u *telejoon.StateUpdate) (telejoon.SwitchAction, telejoon.ShouldPass) {
        log.Printf("User %d in Welcome", u.UserID())
        return telejoon.Continue()
    }).
    OnText(func(c *tgbotapi.TelegramBot, u *telejoon.StateUpdate) (telejoon.SwitchAction, telejoon.ShouldPass) {
        if u.Text() == "secret" {
            return telejoon.GoTo("SecretArea")
        }
        return telejoon.Continue()
    }).
    OnPhoto(func(c *tgbotapi.TelegramBot, u *telejoon.StateUpdate) (telejoon.SwitchAction, telejoon.ShouldPass) {
        c.Send(c.Message().SetChatId(u.ChatID()).SetText("Nice photo!"))
        return telejoon.Stop()
    }).
    Build()

// Register with engine
engine := telejoon.WithPrivateStateHandlers(userRepo, "Welcome").
    AddStaticMenu("Welcome", welcomeMenu)
```

---

## Migration from Old API

| Old API | New API |
|---------|---------|
| `telejoon.NewStaticText("text")` | `text.S("text")` |
| `telejoon.NewLanguageKeyText("key")` | `text.L("key")` |
| `AddStateButton(label, state)` | `buttons.GoTo(label, state)` |
| `AddTextButton(label, response)` | `buttons.Reply(label, response)` |
| `AddConditionalStateButton(cond, ...)` | `buttons.GoTo(...).When(cond)` |
| `return nil, true` | `return telejoon.Continue()` |
| `return NewSwitchActionState("X"), false` | `return telejoon.GoTo("X")` |
| `u.Update.From().Id` | `u.UserID()` |
| `u.Update.Message.Text` | `u.Text()` |
