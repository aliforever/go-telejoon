package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

const (
	dynamicHandlerText      = "TEXT"
	dynamicHandlerPhoto     = "PHOTO"
	dynamicHandlerAudio     = "AUDIO"
	dynamicHandlerDocument  = "DOCUMENT"
	dynamicHandlerSticker   = "STICKER"
	dynamicHandlerVideo     = "VIDEO"
	dynamicHandlerVoice     = "VOICE"
	dynamicHandlerLocation  = "LOCATION"
	dynamicHandlerContact   = "CONTACT"
	dynamicHandlerVideoNote = "VIDEO_NOTE"
	dynamicHandlerVenue     = "VENUE"
	dynamicHandlerPoll      = "POLL"
	dynamicHandlerDice      = "DICE"
)

// DynamicHandlers is a struct that holds all the dynamic handlers
// for example: text, photo, audio, document, sticker, video, voice, location, contact, video_note, venue, poll, dice
// it either returns an empty string or a string that represents the next state
// returning false means the update shouldn't be processed
type DynamicHandlers[User any] struct {
	handlers map[string]DynamicHandlerFunc[User]
}

type DynamicHandler[User any] struct {
	kind    string
	handler DynamicHandlerFunc[User]
}

type DynamicHandlerFunc[User any] func(client *tgbotapi.TelegramBot, update *StateUpdate[User]) (SwitchAction, bool)

// NewDynamicHandlers creates a new DynamicHandlers.
func NewDynamicHandlers[User any](handlers ...*DynamicHandler[User]) *DynamicHandlers[User] {
	dynamicHandlers := &DynamicHandlers[User]{
		handlers: make(map[string]DynamicHandlerFunc[User]),
	}

	for i := range handlers {
		handler := handlers[i]
		if handler != nil {
			dynamicHandlers.handlers[handler.kind] = handler.handler
		}
	}

	return dynamicHandlers
}

// Process processes the given update and returns the next state.
func (d *DynamicHandlers[User]) Process(client *tgbotapi.TelegramBot, update *StateUpdate[User]) (SwitchAction, bool) {
	if update.Update.Message == nil {
		return nil, false
	}

	if update.Update.Message.Text != "" {
		if handler, ok := d.handlers[dynamicHandlerText]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Photo != nil {
		if handler, ok := d.handlers[dynamicHandlerPhoto]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Audio != nil {
		if handler, ok := d.handlers[dynamicHandlerAudio]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Document != nil {
		if handler, ok := d.handlers[dynamicHandlerDocument]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Sticker != nil {
		if handler, ok := d.handlers[dynamicHandlerSticker]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Video != nil {
		if handler, ok := d.handlers[dynamicHandlerVideo]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Voice != nil {
		if handler, ok := d.handlers[dynamicHandlerVoice]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Location != nil {
		if handler, ok := d.handlers[dynamicHandlerLocation]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Contact != nil {
		if handler, ok := d.handlers[dynamicHandlerContact]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.VideoNote != nil {
		if handler, ok := d.handlers[dynamicHandlerVideoNote]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Venue != nil {
		if handler, ok := d.handlers[dynamicHandlerVenue]; ok {
			return handler(client, update)
		}
	}

	if update.Update.Message.Poll != nil {
		if handler, ok := d.handlers[dynamicHandlerPoll]; ok {
			return handler(client, update)
		}
	}

	// if update.Update.Message.Dice != nil {
	// 	if handler, ok := d.handlers[dynamicHandlerDice]; ok {
	// 		return handler(client, update)
	// 	}
	// }

	return nil, false
}

// NewDynamicTextHandler creates a new DynamicHandler[User] with the given text handler.
func NewDynamicTextHandler[User any](textHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerText, handler: textHandler}
}

// NewDynamicPhotoHandler creates a new DynamicHandler[User] with the given photo handler.
func NewDynamicPhotoHandler[User any](photoHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerPhoto, handler: photoHandler}
}

// NewDynamicAudioHandler creates a new DynamicHandler[User] with the given audio handler.
func NewDynamicAudioHandler[User any](audioHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerAudio, handler: audioHandler}
}

// NewDynamicDocumentHandler creates a new DynamicHandler[User] with the given document handler.
func NewDynamicDocumentHandler[User any](documentHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerDocument, handler: documentHandler}
}

// NewDynamicStickerHandler creates a new DynamicHandler[User] with the given sticker handler.
func NewDynamicStickerHandler[User any](stickerHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerSticker, handler: stickerHandler}
}

// NewDynamicVideoHandler creates a new DynamicHandler[User] with the given video handler.
func NewDynamicVideoHandler[User any](videoHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerVideo, handler: videoHandler}
}

// NewDynamicVoiceHandler creates a new DynamicHandler[User] with the given voice handler.
func NewDynamicVoiceHandler[User any](voiceHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerVoice, handler: voiceHandler}
}

// NewDynamicVideoNoteHandler creates a new DynamicHandler[User] with the given video note handler.
func NewDynamicVideoNoteHandler[User any](videoNoteHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerVideoNote, handler: videoNoteHandler}
}

// NewDynamicContactHandler creates a new DynamicHandler[User] with the given contact handler.
func NewDynamicContactHandler[User any](contactHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerContact, handler: contactHandler}
}

// NewDynamicLocationHandler creates a new DynamicHandler[User] with the given location handler.
func NewDynamicLocationHandler[User any](locationHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerLocation, handler: locationHandler}
}

// NewDynamicVenueHandler creates a new DynamicHandler[User] with the given venue handler.
func NewDynamicVenueHandler[User any](venueHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerVenue, handler: venueHandler}
}

// NewDynamicPollHandler creates a new DynamicHandler[User] with the given poll handler.
func NewDynamicPollHandler[User any](pollHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerPoll, handler: pollHandler}
}

// NewDynamicDiceHandler creates a new DynamicHandler[User] with the given dice handler.
func NewDynamicDiceHandler[User any](diceHandler DynamicHandlerFunc[User]) *DynamicHandler[User] {
	return &DynamicHandler[User]{kind: dynamicHandlerDice, handler: diceHandler}
}
