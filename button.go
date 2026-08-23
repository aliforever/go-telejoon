package telejoon

// btnCore holds everything both button families share. Self is the concrete
// button type so modifiers return the right type for continued chaining.
type btnCore[Self any] struct {
	self *Self

	label Text

	whens    []Cond
	unlesses []Cond
	cond     func(ctx *Ctx) bool

	breakBefore bool
	breakAfter  bool

	hook Handler
}

// When shows the button only when the memoized condition evaluates to true.
func (c *btnCore[Self]) When(cond Cond) *Self {
	c.whens = append(c.whens, cond)

	return c.self
}

// Unless shows the button only when the memoized condition evaluates to false.
func (c *btnCore[Self]) Unless(cond Cond) *Self {
	c.unlesses = append(c.unlesses, cond)

	return c.self
}

// If shows the button only when fn returns true for the request.
// Unlike When, fn is not memoized; prefer When for conditions shared between
// buttons.
func (c *btnCore[Self]) If(fn func(ctx *Ctx) bool) *Self {
	c.cond = fn

	return c.self
}

// NewRow starts a new keyboard row before this button.
func (c *btnCore[Self]) NewRow() *Self {
	c.breakBefore = true

	return c.self
}

// Alone puts this button on its own row.
func (c *btnCore[Self]) Alone() *Self {
	c.breakBefore = true
	c.breakAfter = true

	return c.self
}

// Hook runs h before the button's action. If h returns anything other than
// Next, the action is skipped and h's result is processed instead.
func (c *btnCore[Self]) Hook(h Handler) *Self {
	c.hook = h

	return c.self
}

func (c *btnCore[Self]) canBeShown(ctx *Ctx) bool {
	for _, when := range c.whens {
		if !when.eval(ctx) {
			return false
		}
	}

	for _, unless := range c.unlesses {
		if unless.eval(ctx) {
			return false
		}
	}

	if c.cond != nil && !c.cond(ctx) {
		return false
	}

	return true
}

// === Reply keyboard buttons ===

type buttonKind int

const (
	buttonKindText buttonKind = iota
	buttonKindState
	buttonKindInlineMenu
	buttonKindRaw
)

// Button is a reply-keyboard button. Construct with GoTo, GoToWith, Reply,
// ShowInline, or Raw, then apply modifiers (When, Unless, If, NewRow, Alone,
// Hook).
type Button struct {
	btnCore[Button]

	kind buttonKind

	text  Text   // response text (buttonKindText)
	state string // target state (buttonKindState)
	data  any    // *D payload (buttonKindState via GoToWith)
	menu  string // target inline menu (buttonKindInlineMenu)
}

func newButton(kind buttonKind, label Text) *Button {
	button := &Button{kind: kind}
	button.self = button
	button.label = label

	return button
}

// GoTo creates a button that transitions the user to a payload-less state.
func GoTo(label Text, state State[NoData]) *Button {
	button := newButton(buttonKindState, label)
	button.state = state.name

	return button
}

// GoToWith creates a button that transitions the user to a state carrying a
// typed payload. D is inferred from the state handle.
func GoToWith[D any](label Text, state State[D], data D) *Button {
	button := newButton(buttonKindState, label)
	button.state = state.name
	button.data = &data

	return button
}

// Reply creates a button that sends a text response when pressed.
func Reply(label Text, response Text) *Button {
	button := newButton(buttonKindText, label)
	button.text = response

	return button
}

// ShowInline creates a button that sends the given inline menu when pressed.
func ShowInline(label Text, menu InlineMenuRef) *Button {
	button := newButton(buttonKindInlineMenu, label)
	button.menu = menu.name

	return button
}

// Raw creates a button with no built-in action. It renders like any other
// button; when pressed, the message falls through to the menu's OnText
// handler, which can match on ctx.Text().
func Raw(label Text) *Button {
	return newButton(buttonKindRaw, label)
}

// === Inline keyboard buttons ===

type inlineKind int

const (
	inlineKindURL inlineKind = iota
	inlineKindMenu
	inlineKindAlert
	inlineKindState
	inlineKindCallback
)

// InlineButton is an inline-keyboard button. Construct with Do, URL, Alert,
// AlertDialog, OpenMenu, OpenMenuEdit, or StateBtn, then apply modifiers.
type InlineButton struct {
	btnCore[InlineButton]

	kind inlineKind

	url Text // inlineKindURL

	alertText string // inlineKindAlert
	showAlert bool

	menu string // inlineKindMenu
	edit bool

	state string // inlineKindState

	callbackData string // inlineKindCallback (pre-encoded by Do)
	encodeErr    error
}

func newInlineButton(kind inlineKind, label Text) *InlineButton {
	button := &InlineButton{kind: kind}
	button.self = button
	button.label = label

	return button
}

// Do creates an inline button that invokes the given typed route with the
// given payload. The payload is encoded with the route's codec, so the
// button-side encoding and the handler-side decoding can never drift:
//
//	del := products.Route("del", func(ctx *telejoon.Ctx, a DelArgs) telejoon.Action { ... })
//	telejoon.Do(telejoon.S("🗑"), del, DelArgs{ProductID: p.ID})
func Do[A any](label Text, route Route[A], args A) *InlineButton {
	button := newInlineButton(inlineKindCallback, label)
	button.callbackData, button.encodeErr = route.encode(args)

	return button
}

// URL creates an inline button that opens a URL.
func URL(label Text, url Text) *InlineButton {
	button := newInlineButton(inlineKindURL, label)
	button.url = url

	return button
}

// Alert creates an inline button that shows a toast notification.
func Alert(label Text, text string) *InlineButton {
	button := newInlineButton(inlineKindAlert, label)
	button.alertText = text

	return button
}

// AlertDialog creates an inline button that shows a popup dialog.
func AlertDialog(label Text, text string) *InlineButton {
	button := newInlineButton(inlineKindAlert, label)
	button.alertText = text
	button.showAlert = true

	return button
}

// OpenMenu creates an inline button that sends another inline menu.
func OpenMenu(label Text, menu InlineMenuRef) *InlineButton {
	button := newInlineButton(inlineKindMenu, label)
	button.menu = menu.name

	return button
}

// OpenMenuEdit creates an inline button that edits the current message into
// another inline menu.
func OpenMenuEdit(label Text, menu InlineMenuRef) *InlineButton {
	button := newInlineButton(inlineKindMenu, label)
	button.menu = menu.name
	button.edit = true

	return button
}

// StateBtn creates an inline button that transitions the user to a
// payload-less state.
func StateBtn(label Text, state State[NoData]) *InlineButton {
	button := newInlineButton(inlineKindState, label)
	button.state = state.name

	return button
}
