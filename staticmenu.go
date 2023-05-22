package telejoon

import (
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

// StaticMenu is a Handler that receives predefined handlers and acts accordingly.
type StaticMenu struct {
	lock sync.Mutex

	textBuilder TextBuilder

	actionBuilder ActionBuilderKind

	deferredActionBuilder DeferredActionBuilder

	dynamicHandlers map[string]Handler

	middlewares []Middleware
}

type (
	StaticMenuMiddleware func(*tgbotapi.TelegramBot, *StateUpdate) (SwitchAction, bool)
)

func parseMiddlewaresAndDynamicHandlers(handlers ...Handler) (
	[]Middleware, map[string]Handler,
) {

	var middlewares []Middleware
	dynamicHandlers := map[string]Handler{}

	for _, handler := range handlers {
		switch h := handler.(type) {
		case Middleware:
			middlewares = append(middlewares, h)
		case DynamicHandlerText:
			dynamicHandlers[TextHandler] = h
		case DynamicHandlerPhoto:
			dynamicHandlers[PhotoHandler] = h
		case DynamicHandlerVideo:
			dynamicHandlers[VideoHandler] = h
		case DynamicHandlerVoice:
			dynamicHandlers[VoiceHandler] = h
		case DynamicHandlerAudio:
			dynamicHandlers[AudioHandler] = h
		case DynamicHandlerDocument:
			dynamicHandlers[DocumentHandler] = h
		case DynamicHandlerSticker:
			dynamicHandlers[StickerHandler] = h
		case DynamicHandlerLocation:
			dynamicHandlers[LocationHandler] = h
		case DynamicHandlerContact:
			dynamicHandlers[ContactHandler] = h
		case DynamicHandlerVideoNote:
			dynamicHandlers[VideoNoteHandler] = h
		case DynamicHandlerVenue:
			dynamicHandlers[VenueHandler] = h
		case DynamicHandlerPoll:
			dynamicHandlers[PollHandler] = h
		case DynamicHandlerDice:
			dynamicHandlers[DiceHandler] = h
		case DynamicHandler:
			dynamicHandlers[DefaultHandler] = h
		default:
			panic(fmt.Sprintf("invalid Handler type: %T", handler))
		}
	}

	return middlewares, dynamicHandlers
}

// NewStaticMenu creates a new StaticMenu with the given text and action builder.
func NewStaticMenu(
	text TextBuilder,
	builder ActionBuilderKind,
	middlewaresAndDynamicHandlers ...Handler) *StaticMenu {

	middlewares, handlers := parseMiddlewaresAndDynamicHandlers(middlewaresAndDynamicHandlers...)

	return &StaticMenu{
		textBuilder:     text,
		actionBuilder:   builder,
		middlewares:     middlewares,
		dynamicHandlers: handlers,
	}
}

// processReplyText with StateUpdate and returns the text to be replied.
func (s *StaticMenu) processReplyText(update *StateUpdate) string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.textBuilder.String(update)
}

func (s *StaticMenu) processActionBuilder(update *StateUpdate) *ActionBuilder {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.actionBuilder.Build(update)
}
