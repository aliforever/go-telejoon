package telejoon

import (
	"fmt"
	"sync"
	"sync/atomic"

	tgbotapi "github.com/aliforever/go-telegram-bot-api"
)

// Key is a typed session-storage key. Keys are unforgeable: two keys created
// with the same name (or different type parameters) never collide, because
// identity is a unique sequence number, not the name. The name is only for
// debugging.
//
//	var CartKey = telejoon.NewKey[[]CartItem]("cart")
type Key[T any] struct {
	id   uint64
	name string
}

var keyCounter atomic.Uint64

// NewKey creates a new typed session-storage key.
func NewKey[T any](name string) Key[T] {
	return Key[T]{id: keyCounter.Add(1), name: name}
}

// String returns the debug name of the key.
func (k Key[T]) String() string {
	return k.name
}

// Ctx is the per-update request context: it carries the Telegram client, the
// raw update, the current state, the user's language, typed session storage,
// and the resolved state payload.
//
// A Ctx belongs to the goroutine processing its update. Spawned child
// goroutines must synchronize their own access.
type Ctx struct {
	client *tgbotapi.TelegramBot
	engine *Engine

	session *sync.Map

	// Update is the raw Telegram update being processed.
	Update tgbotapi.Update

	// State is the name of the user's current state.
	State string

	language *Language

	// IsSwitched reports whether the state was switched during this request.
	IsSwitched bool

	condResults map[uint64]bool

	// userID is set when the Ctx is built without an update carrying a
	// sender (e.g. SwitchUserState), and takes precedence in UserID/ChatID.
	userID int64

	// switchDepth guards against unbounded middleware redirect loops.
	switchDepth int

	// pendingData holds a *D carried by a GoToWith transition made during this
	// request, before it is resolved into stateData for the target menu.
	pendingData any

	// stateData holds a *D resolved for the current state's menu.
	stateData any
}

// === Typed session storage (Go 1.27 generic methods) ===

// Set stores value under key. The value's type is checked against the key's
// type parameter at compile time.
func (c *Ctx) Set[T any](key Key[T], value T) {
	c.session.Store(key.id, value)
}

// Get loads the value stored under key. T is inferred from the key, so call
// sites need no type annotation:
//
//	cart, ok := ctx.Get(CartKey) // cart is []CartItem
func (c *Ctx) Get[T any](key Key[T]) (T, bool) {
	var zero T

	value, ok := c.session.Load(key.id)
	if !ok {
		return zero, false
	}

	typed, ok := value.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

// GetOr loads the value stored under key, or returns def when absent or of a
// mismatched type.
func (c *Ctx) GetOr[T any](key Key[T], def T) T {
	if value, ok := c.Get(key); ok {
		return value
	}

	return def
}

// Has reports whether key holds a value.
func (c *Ctx) Has[T any](key Key[T]) bool {
	_, ok := c.session.Load(key.id)

	return ok
}

// Delete removes the value stored under key.
func (c *Ctx) Delete[T any](key Key[T]) {
	c.session.Delete(key.id)
}

// === Client, language, and raw access ===

// Client returns the underlying Telegram bot client for advanced use.
func (c *Ctx) Client() *tgbotapi.TelegramBot {
	return c.client
}

// Language returns the user's resolved language, or nil when languages are
// not configured or the user has not chosen one yet.
func (c *Ctx) Language() *Language {
	return c.language
}

// SetLanguage overrides the language for this request.
func (c *Ctx) SetLanguage(language *Language) {
	c.language = language
}

// === Transitions and responses ===

// GoTo transitions the user to a payload-less state.
func (c *Ctx) GoTo(state State[NoData]) Action {
	return actionResult{kind: actionKindState, state: state.name}
}

// GoToWith transitions the user to a state carrying a typed payload. D is
// inferred from the state handle; passing the wrong payload type is a compile
// error.
func (c *Ctx) GoToWith[D any](state State[D], data D) Action {
	return actionResult{kind: actionKindState, state: state.name, data: &data}
}

// ReplyText sends a text message to the user and stops processing.
func (c *Ctx) ReplyText(text string) Action {
	_, err := c.client.Send(c.client.Message().SetText(text).SetChatId(c.ChatID()))
	if err != nil {
		return Error(fmt.Errorf("error_sending_message_to_user: %d, %w", c.ChatID(), err))
	}

	return Stop()
}

// AnswerCallback answers the current callback query with an optional toast or
// alert, and stops processing.
func (c *Ctx) AnswerCallback(text string, showAlert bool) Action {
	if c.Update.CallbackQuery == nil {
		return Error(fmt.Errorf("answer_callback: not a callback query"))
	}

	_, err := c.client.Send(c.client.AnswerCallbackQuery().
		SetCallbackQueryId(c.Update.CallbackQuery.Id).
		SetText(text).
		SetShowAlert(showAlert))
	if err != nil {
		return Error(fmt.Errorf("answer_callback: %w", err))
	}

	return Stop()
}

// === Convenience accessors ===

// UserID returns the user ID from the update (or the explicit user the Ctx
// was built for, e.g. via SwitchUserState).
func (c *Ctx) UserID() int64 {
	if c.userID != 0 {
		return c.userID
	}

	if from := c.Update.From(); from != nil {
		return from.Id
	}

	return 0
}

// ChatID returns the chat ID from the update, falling back to the explicit
// user ID (private chats share chat and user IDs).
func (c *Ctx) ChatID() int64 {
	if chat := c.Update.Chat(); chat != nil {
		return chat.Id
	}

	return c.userID
}

// Username returns the username from the update (without @).
func (c *Ctx) Username() string {
	if from := c.Update.From(); from != nil {
		return from.Username
	}

	return ""
}

// FirstName returns the first name from the update.
func (c *Ctx) FirstName() string {
	if from := c.Update.From(); from != nil {
		return from.FirstName
	}

	return ""
}

// Text returns the message text, or empty string if not a text message.
func (c *Ctx) Text() string {
	if c.Update.Message != nil {
		return c.Update.Message.Text
	}

	return ""
}

// CallbackData returns the callback query data, or empty string.
func (c *Ctx) CallbackData() string {
	if c.Update.CallbackQuery != nil {
		return c.Update.CallbackQuery.Data
	}

	return ""
}

// CallbackID returns the callback query ID, or empty string.
func (c *Ctx) CallbackID() string {
	if c.Update.CallbackQuery != nil {
		return c.Update.CallbackQuery.Id
	}

	return ""
}

// Command parses and returns the command from the message, or nil if not a command.
func (c *Ctx) Command() *Command {
	if c.Update.Message == nil || c.Update.Message.Text == "" {
		return nil
	}

	return ParseCommand(c.Update.Message.Text)
}

// IsCommand returns true if the message is a command (starts with /).
func (c *Ctx) IsCommand() bool {
	if c.Update.Message == nil {
		return false
	}

	return IsCommand(c.Update.Message.Text)
}

// === Message type detection ===

// IsText returns true if this is a text message.
func (c *Ctx) IsText() bool {
	return c.Update.Message != nil && c.Update.Message.Text != ""
}

// IsPhoto returns true if this is a photo message.
func (c *Ctx) IsPhoto() bool {
	return c.Update.Message != nil && c.Update.Message.Photo != nil
}

// IsDocument returns true if this is a document message.
func (c *Ctx) IsDocument() bool {
	return c.Update.Message != nil && c.Update.Message.Document != nil
}

// IsVideo returns true if this is a video message.
func (c *Ctx) IsVideo() bool {
	return c.Update.Message != nil && c.Update.Message.Video != nil
}

// IsVoice returns true if this is a voice message.
func (c *Ctx) IsVoice() bool {
	return c.Update.Message != nil && c.Update.Message.Voice != nil
}

// IsAudio returns true if this is an audio message.
func (c *Ctx) IsAudio() bool {
	return c.Update.Message != nil && c.Update.Message.Audio != nil
}

// IsLocation returns true if this is a location message.
func (c *Ctx) IsLocation() bool {
	return c.Update.Message != nil && c.Update.Message.Location != nil
}

// IsContact returns true if this is a contact message.
func (c *Ctx) IsContact() bool {
	return c.Update.Message != nil && c.Update.Message.Contact != nil
}

// IsSticker returns true if this is a sticker message.
func (c *Ctx) IsSticker() bool {
	return c.Update.Message != nil && c.Update.Message.Sticker != nil
}

// IsCallback returns true if this is a callback query.
func (c *Ctx) IsCallback() bool {
	return c.Update.CallbackQuery != nil
}

// MessageType returns the type of message as a string.
// Returns one of: "text", "photo", "document", "video", "voice", "audio",
// "location", "contact", "sticker", "callback", or "unknown".
func (c *Ctx) MessageType() string {
	switch {
	case c.IsCallback():
		return "callback"
	case c.IsPhoto():
		return "photo"
	case c.IsDocument():
		return "document"
	case c.IsVideo():
		return "video"
	case c.IsVoice():
		return "voice"
	case c.IsAudio():
		return "audio"
	case c.IsLocation():
		return "location"
	case c.IsContact():
		return "contact"
	case c.IsSticker():
		return "sticker"
	case c.IsText():
		return "text"
	default:
		return "unknown"
	}
}
