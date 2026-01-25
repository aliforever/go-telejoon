package telejoon

import (
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api"
)

type StateUpdate struct {
	storage *sync.Map

	State      string
	language   *Language
	Update     tgbotapi.Update
	IsSwitched bool
}

// Set sets a value for the context.
func (s *StateUpdate) Set(key, value interface{}) {
	s.storage.Store(key, value)
}

// Get gets a value from the context.
func (s *StateUpdate) Get(key interface{}) interface{} {
	value, _ := s.storage.Load(key)

	return value
}

// SetLanguage sets the language for the user.
func (s *StateUpdate) SetLanguage(language *Language) {
	s.language = language
}

func (s *StateUpdate) Language() *Language {
	return s.language
}

// Command parses and returns the command from the message, or nil if not a command.
// This is a convenience method that calls ParseCommand on the message text.
func (s *StateUpdate) Command() *Command {
	if s.Update.Message == nil || s.Update.Message.Text == "" {
		return nil
	}
	return ParseCommand(s.Update.Message.Text)
}

// IsCommand returns true if the message is a command (starts with /).
func (s *StateUpdate) IsCommand() bool {
	if s.Update.Message == nil {
		return false
	}
	return IsCommand(s.Update.Message.Text)
}

// === Typed Getters ===

// GetString gets a string value from storage with a default.
func (s *StateUpdate) GetString(key string, defaultVal string) string {
	if value, ok := s.storage.Load(key); ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultVal
}

// GetInt gets an int value from storage with a default.
func (s *StateUpdate) GetInt(key string, defaultVal int) int {
	if value, ok := s.storage.Load(key); ok {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return defaultVal
}

// GetInt64 gets an int64 value from storage with a default.
func (s *StateUpdate) GetInt64(key string, defaultVal int64) int64 {
	if value, ok := s.storage.Load(key); ok {
		if i, ok := value.(int64); ok {
			return i
		}
	}
	return defaultVal
}

// GetBool gets a bool value from storage with a default.
func (s *StateUpdate) GetBool(key string, defaultVal bool) bool {
	if value, ok := s.storage.Load(key); ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// Has returns true if the key exists in storage.
func (s *StateUpdate) Has(key string) bool {
	_, ok := s.storage.Load(key)
	return ok
}

// Delete removes a key from storage.
func (s *StateUpdate) Delete(key string) {
	s.storage.Delete(key)
}

// Clear removes all keys from storage.
func (s *StateUpdate) Clear() {
	s.storage.Range(func(key, value interface{}) bool {
		s.storage.Delete(key)
		return true
	})
}

// === Convenience Accessors ===

// UserID returns the user ID from the update.
func (s *StateUpdate) UserID() int64 {
	if from := s.Update.From(); from != nil {
		return from.Id
	}
	return 0
}

// ChatID returns the chat ID from the update.
func (s *StateUpdate) ChatID() int64 {
	if chat := s.Update.Chat(); chat != nil {
		return chat.Id
	}
	return 0
}

// Username returns the username from the update (without @).
func (s *StateUpdate) Username() string {
	if from := s.Update.From(); from != nil {
		return from.Username
	}
	return ""
}

// FirstName returns the first name from the update.
func (s *StateUpdate) FirstName() string {
	if from := s.Update.From(); from != nil {
		return from.FirstName
	}
	return ""
}

// Text returns the message text, or empty string if not a text message.
func (s *StateUpdate) Text() string {
	if s.Update.Message != nil {
		return s.Update.Message.Text
	}
	return ""
}

// CallbackData returns the callback query data, or empty string.
func (s *StateUpdate) CallbackData() string {
	if s.Update.CallbackQuery != nil {
		return s.Update.CallbackQuery.Data
	}
	return ""
}

// CallbackID returns the callback query ID, or empty string.
func (s *StateUpdate) CallbackID() string {
	if s.Update.CallbackQuery != nil {
		return s.Update.CallbackQuery.Id
	}
	return ""
}

// === Message Type Detection ===

// IsText returns true if this is a text message.
func (s *StateUpdate) IsText() bool {
	return s.Update.Message != nil && s.Update.Message.Text != ""
}

// IsPhoto returns true if this is a photo message.
func (s *StateUpdate) IsPhoto() bool {
	return s.Update.Message != nil && s.Update.Message.Photo != nil
}

// IsDocument returns true if this is a document message.
func (s *StateUpdate) IsDocument() bool {
	return s.Update.Message != nil && s.Update.Message.Document != nil
}

// IsVideo returns true if this is a video message.
func (s *StateUpdate) IsVideo() bool {
	return s.Update.Message != nil && s.Update.Message.Video != nil
}

// IsVoice returns true if this is a voice message.
func (s *StateUpdate) IsVoice() bool {
	return s.Update.Message != nil && s.Update.Message.Voice != nil
}

// IsAudio returns true if this is an audio message.
func (s *StateUpdate) IsAudio() bool {
	return s.Update.Message != nil && s.Update.Message.Audio != nil
}

// IsLocation returns true if this is a location message.
func (s *StateUpdate) IsLocation() bool {
	return s.Update.Message != nil && s.Update.Message.Location != nil
}

// IsContact returns true if this is a contact message.
func (s *StateUpdate) IsContact() bool {
	return s.Update.Message != nil && s.Update.Message.Contact != nil
}

// IsSticker returns true if this is a sticker message.
func (s *StateUpdate) IsSticker() bool {
	return s.Update.Message != nil && s.Update.Message.Sticker != nil
}

// IsCallback returns true if this is a callback query.
func (s *StateUpdate) IsCallback() bool {
	return s.Update.CallbackQuery != nil
}

// MessageType returns the type of message as a string.
// Returns one of: "text", "photo", "document", "video", "voice", "audio",
// "location", "contact", "sticker", "callback", or "unknown".
func (s *StateUpdate) MessageType() string {
	switch {
	case s.IsCallback():
		return "callback"
	case s.IsPhoto():
		return "photo"
	case s.IsDocument():
		return "document"
	case s.IsVideo():
		return "video"
	case s.IsVoice():
		return "voice"
	case s.IsAudio():
		return "audio"
	case s.IsLocation():
		return "location"
	case s.IsContact():
		return "contact"
	case s.IsSticker():
		return "sticker"
	case s.IsText():
		return "text"
	default:
		return "unknown"
	}
}
