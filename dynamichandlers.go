package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

// DynamicHandlers is a struct that holds all the dynamic handlers
// for example: text, photo, audio, document, sticker, video, voice, location, contact, video_note, venue, poll, dice
// it either returns an empty string or a string that represents the next state
// returning false means the update shouldn't be processed
type dynamicHandlers[User any] struct {
	textHandler      DynamicHandler[User]
	photoHandler     DynamicHandler[User]
	audioHandler     DynamicHandler[User]
	documentHandler  DynamicHandler[User]
	stickerHandler   DynamicHandler[User]
	videoHandler     DynamicHandler[User]
	voiceHandler     DynamicHandler[User]
	locationHandler  DynamicHandler[User]
	contactHandler   DynamicHandler[User]
	videoNoteHandler DynamicHandler[User]
	venueHandler     DynamicHandler[User]
	pollHandler      DynamicHandler[User]
	diceHandler      DynamicHandler[User]
}

type (
	DynamicHandler[User any] func(client *tgbotapi.TelegramBot, update *StateUpdate[User]) (SwitchAction, bool)
)

// NewDynamicHandlers creates a new DynamicHandlers.
func NewDynamicHandlers[User any]() *dynamicHandlers[User] {
	return &dynamicHandlers[User]{}
}

// WithTextHandler sets the textHandler handler.
func (d *dynamicHandlers[User]) WithTextHandler(
	textHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.textHandler = textHandler

	return d
}

// WithPhotoHandler sets the photoHandler handler.
func (d *dynamicHandlers[User]) WithPhotoHandler(
	photoHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.photoHandler = photoHandler

	return d
}

// WithAudioHandler sets the audioHandler handler.
func (d *dynamicHandlers[User]) WithAudioHandler(
	audioHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.audioHandler = audioHandler

	return d
}

// WithDocumentHandler sets the documentHandler handler.
func (d *dynamicHandlers[User]) WithDocumentHandler(
	documentHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.documentHandler = documentHandler

	return d
}

// WithStickerHandler sets the stickerHandler handler.
func (d *dynamicHandlers[User]) WithStickerHandler(
	stickerHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.stickerHandler = stickerHandler

	return d
}

// WithVideoHandler sets the videoHandler handler.
func (d *dynamicHandlers[User]) WithVideoHandler(
	videoHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.videoHandler = videoHandler

	return d
}

// WithVoiceHandler sets the voiceHandler handler.
func (d *dynamicHandlers[User]) WithVoiceHandler(
	voiceHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.voiceHandler = voiceHandler

	return d
}

// WithVideoNoteHandler sets the videoNoteHandler handler.
func (d *dynamicHandlers[User]) WithVideoNoteHandler(
	videoNoteHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.videoNoteHandler = videoNoteHandler

	return d
}

// WithContactHandler sets the contactHandler handler.
func (d *dynamicHandlers[User]) WithContactHandler(
	contactHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.contactHandler = contactHandler

	return d
}

// WithLocationHandler sets the locationHandler handler.
func (d *dynamicHandlers[User]) WithLocationHandler(
	locationHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.locationHandler = locationHandler

	return d
}

// WithVenueHandler sets the venueHandler handler.
func (d *dynamicHandlers[User]) WithVenueHandler(
	venueHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.venueHandler = venueHandler

	return d
}

// WithPollHandler sets the pollHandler handler.
func (d *dynamicHandlers[User]) WithPollHandler(
	pollHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.pollHandler = pollHandler

	return d
}

// WithDiceHandler sets the diceHandler handler.
func (d *dynamicHandlers[User]) WithDiceHandler(
	diceHandler DynamicHandler[User]) *dynamicHandlers[User] {

	d.diceHandler = diceHandler

	return d
}
