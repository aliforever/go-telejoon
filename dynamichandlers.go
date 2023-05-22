package telejoon

type DynamicHandler struct {
	UpdateHandler
}

const (
	DefaultHandler   = "DEFAULT"
	TextHandler      = "TEXT"
	PhotoHandler     = "PHOTO"
	AudioHandler     = "AUDIO"
	DocumentHandler  = "DOCUMENT"
	StickerHandler   = "STICKER"
	VideoHandler     = "VIDEO"
	VoiceHandler     = "VOICE"
	LocationHandler  = "LOCATION"
	ContactHandler   = "CONTACT"
	VideoNoteHandler = "VIDEO_NOTE"
	VenueHandler     = "VENUE"
	PollHandler      = "POLL"
	DiceHandler      = "DICE"
)

type DynamicHandlerText struct {
	UpdateHandler
}

type DynamicHandlerPhoto struct {
	UpdateHandler
}

type DynamicHandlerAudio struct {
	UpdateHandler
}

type DynamicHandlerDocument struct {
	UpdateHandler
}

type DynamicHandlerSticker struct {
	UpdateHandler
}

type DynamicHandlerVideo struct {
	UpdateHandler
}

type DynamicHandlerVoice struct {
	UpdateHandler
}

type DynamicHandlerLocation struct {
	UpdateHandler
}

type DynamicHandlerContact struct {
	UpdateHandler
}

type DynamicHandlerVideoNote struct {
	UpdateHandler
}

type DynamicHandlerVenue struct {
	UpdateHandler
}

type DynamicHandlerPoll struct {
	UpdateHandler
}

type DynamicHandlerDice struct {
	UpdateHandler
}

// NewDynamicHandlerText creates a new DynamicHandlerText
func NewDynamicHandlerText(handler UpdateHandler) Handler {
	return DynamicHandlerText{UpdateHandler: handler}
}

// NewDynamicHandlerPhoto creates a new DynamicHandlerPhoto
func NewDynamicHandlerPhoto(handler UpdateHandler) Handler {
	return DynamicHandlerPhoto{UpdateHandler: handler}
}

// NewDynamicHandlerAudio creates a new DynamicHandlerAudio
func NewDynamicHandlerAudio(handler UpdateHandler) Handler {
	return DynamicHandlerAudio{UpdateHandler: handler}
}

// NewDynamicHandlerDocument creates a new DynamicHandlerDocument
func NewDynamicHandlerDocument(handler UpdateHandler) Handler {
	return DynamicHandlerDocument{UpdateHandler: handler}
}

// NewDynamicHandlerSticker creates a new DynamicHandlerSticker
func NewDynamicHandlerSticker(handler UpdateHandler) Handler {
	return DynamicHandlerSticker{UpdateHandler: handler}
}

// NewDynamicHandlerVideo creates a new DynamicHandlerVideo
func NewDynamicHandlerVideo(handler UpdateHandler) Handler {
	return DynamicHandlerVideo{UpdateHandler: handler}
}

// NewDynamicHandlerVoice creates a new DynamicHandlerVoice
func NewDynamicHandlerVoice(handler UpdateHandler) Handler {
	return DynamicHandlerVoice{UpdateHandler: handler}
}

// NewDynamicHandlerLocation creates a new DynamicHandlerLocation
func NewDynamicHandlerLocation(handler UpdateHandler) Handler {
	return DynamicHandlerLocation{UpdateHandler: handler}
}

// NewDynamicHandlerContact creates a new DynamicHandlerContact
func NewDynamicHandlerContact(handler UpdateHandler) Handler {
	return DynamicHandlerContact{UpdateHandler: handler}
}

// NewDynamicHandlerVideoNote creates a new DynamicHandlerVideoNote
func NewDynamicHandlerVideoNote(handler UpdateHandler) Handler {
	return DynamicHandlerVideoNote{UpdateHandler: handler}
}

// NewDynamicHandlerVenue creates a new DynamicHandlerVenue
func NewDynamicHandlerVenue(handler UpdateHandler) Handler {
	return DynamicHandlerVenue{UpdateHandler: handler}
}

// NewDynamicHandlerPoll creates a new DynamicHandlerPoll
func NewDynamicHandlerPoll(handler UpdateHandler) Handler {
	return DynamicHandlerPoll{UpdateHandler: handler}
}

// NewDynamicHandlerDice creates a new DynamicHandlerDice
func NewDynamicHandlerDice(handler UpdateHandler) Handler {
	return DynamicHandlerDice{UpdateHandler: handler}
}

// NewDefaultHandler creates a new DefaultHandler
func NewDefaultHandler(handler UpdateHandler) Handler {
	return DynamicHandler{UpdateHandler: handler}
}
