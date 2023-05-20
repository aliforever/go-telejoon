package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

// StaticMenu is a handler that receives predefined handlers and acts accordingly.
type StaticMenu[User any] struct {
	lock sync.Mutex

	replyText            string
	replyWithLanguageKey string
	deferredReplyText    DeferredTextBuilder[User]

	actionBuilder         *ActionBuilder
	deferredActionBuilder DeferredActionBuilder[User]

	dynamicHandlers *DynamicHandlers[User]

	middlewares []StaticMenuMiddleware[User]
}

type (
	StaticMenuMiddleware[User any]  func(*tgbotapi.TelegramBot, *StateUpdate[User]) (SwitchAction, bool)
	DeferredActionBuilder[User any] func(bot *tgbotapi.TelegramBot, update *StateUpdate[User]) *ActionBuilder
	DeferredTextBuilder[User any]   func(update *StateUpdate[User]) string
)

// NewStaticMenuWithTextAndActionBuilder creates a new StaticMenu[User] with the given text and action builder.
func NewStaticMenuWithTextAndActionBuilder[User any](
	text string,
	builder *ActionBuilder,
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyText:     text,
		actionBuilder: builder,
		middlewares:   middlewares,
	}
}

// NewStaticMenuWithTextAndActionBuilderAndDynamicHandlers creates a new StaticMenu[User] with the given text and action
// builder and dynamic handlers.
func NewStaticMenuWithTextAndActionBuilderAndDynamicHandlers[User any](
	text string,
	builder *ActionBuilder,
	dynamicHandlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyText:       text,
		actionBuilder:   builder,
		dynamicHandlers: dynamicHandlers,
		middlewares:     middlewares,
	}
}

// NewStaticMenuWithLanguageKeyAndActionBuilder creates a new StaticMenu[User] with the given language key and
// action builder.
func NewStaticMenuWithLanguageKeyAndActionBuilder[User any](
	key string,
	builder *ActionBuilder,
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyWithLanguageKey: key,
		actionBuilder:        builder,
		middlewares:          middlewares,
	}
}

// NewStaticMenuWithLanguageKeyAndActionBuilderAndDynamicHandlers creates a new StaticMenu[User] with the given
// language key and action builder and dynamic handlers.
func NewStaticMenuWithLanguageKeyAndActionBuilderAndDynamicHandlers[User any](
	key string,
	builder *ActionBuilder,
	dynamicHandlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyWithLanguageKey: key,
		actionBuilder:        builder,
		dynamicHandlers:      dynamicHandlers,
		middlewares:          middlewares,
	}
}

// NewStaticMenuWithTextAndDeferredActionBuilder creates a new StaticMenu[User] with the given text and deferred
// action builder.
func NewStaticMenuWithTextAndDeferredActionBuilder[User any](
	text string,
	deferredBuilder DeferredActionBuilder[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyText:             text,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

// NewStaticMenuWithTextAndDeferredActionBuilderAndDynamicHandlers creates a new StaticMenu[User] with the given text
// and deferred action builder and dynamic handlers.
func NewStaticMenuWithTextAndDeferredActionBuilderAndDynamicHandlers[User any](
	text string,
	deferredBuilder DeferredActionBuilder[User],
	dynamicHandlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyText:             text,
		deferredActionBuilder: deferredBuilder,
		dynamicHandlers:       dynamicHandlers,
		middlewares:           middlewares,
	}
}

// NewStaticMenuWithLanguageKeyAndDeferredActionBuilder creates a new StaticMenu[User] with the given language key
// and deferred action builder.
func NewStaticMenuWithLanguageKeyAndDeferredActionBuilder[User any](
	key string,
	deferredBuilder DeferredActionBuilder[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyWithLanguageKey:  key,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

// NewStaticMenuWithLanguageKeyAndDeferredActionBuilderAndDynamicHandlers creates a new StaticMenu[User] with the
// given language key and deferred action builder and dynamic handlers.
func NewStaticMenuWithLanguageKeyAndDeferredActionBuilderAndDynamicHandlers[User any](
	key string,
	deferredBuilder DeferredActionBuilder[User],
	dynamicHandlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyWithLanguageKey:  key,
		deferredActionBuilder: deferredBuilder,
		dynamicHandlers:       dynamicHandlers,
		middlewares:           middlewares,
	}
}

// NewStaticMenuWithDeferredTextAndActionBuilder creates a new StaticMenu[User] with the given deferred text and
// action builder.
func NewStaticMenuWithDeferredTextAndActionBuilder[User any](
	deferredText DeferredTextBuilder[User],
	builder *ActionBuilder,
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		deferredReplyText: deferredText,
		actionBuilder:     builder,
		middlewares:       middlewares,
	}
}

// NewStaticMenuWithDeferredTextAndActionBuilderAndDynamicHandlers creates a new StaticMenu[User] with the given
// deferred text and action builder and dynamic handlers.
func NewStaticMenuWithDeferredTextAndActionBuilderAndDynamicHandlers[User any](
	deferredText DeferredTextBuilder[User],
	builder *ActionBuilder,
	dynamicHandlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		deferredReplyText: deferredText,
		actionBuilder:     builder,
		dynamicHandlers:   dynamicHandlers,
		middlewares:       middlewares,
	}
}

// NewStaticMenuWithDeferredTextAndDeferredActionBuilder creates a new StaticMenu[User] with the given deferred text
// and deferred action builder.
func NewStaticMenuWithDeferredTextAndDeferredActionBuilder[User any](
	deferredText DeferredTextBuilder[User],
	deferredBuilder DeferredActionBuilder[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		deferredReplyText:     deferredText,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

// NewStaticMenuWithDeferredTextAndDeferredActionBuilderAndDynamicHandlers creates a new StaticMenu[User] with the
// given deferred text and deferred action builder and dynamic handlers.
func NewStaticMenuWithDeferredTextAndDeferredActionBuilderAndDynamicHandlers[User any](
	deferredText DeferredTextBuilder[User],
	deferredBuilder DeferredActionBuilder[User],
	dynamicHandlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		deferredReplyText:     deferredText,
		deferredActionBuilder: deferredBuilder,
		dynamicHandlers:       dynamicHandlers,
		middlewares:           middlewares,
	}
}

// NewStaticMenuWithTextAndDynamicHandlers creates a new StaticMenu[User] with the given text and dynamic handlers.
func NewStaticMenuWithTextAndDynamicHandlers[User any](
	text string,
	handlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyText:       text,
		dynamicHandlers: handlers,
		middlewares:     middlewares,
	}
}

// NewStaticMenuWithLanguageKeyAndDynamicHandlers creates a new StaticMenu[User] with the given language key and
// dynamic handlers.
func NewStaticMenuWithLanguageKeyAndDynamicHandlers[User any](
	key string,
	handlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		replyWithLanguageKey: key,
		dynamicHandlers:      handlers,
		middlewares:          middlewares,
	}
}

// NewStaticMenuWithDeferredTextAndDynamicHandlers creates a new StaticMenu[User] with the given deferred text and
// dynamic handlers.
func NewStaticMenuWithDeferredTextAndDynamicHandlers[User any](
	deferredText DeferredTextBuilder[User],
	handlers *DynamicHandlers[User],
	middlewares ...StaticMenuMiddleware[User]) *StaticMenu[User] {

	return &StaticMenu[User]{
		deferredReplyText: deferredText,
		dynamicHandlers:   handlers,
		middlewares:       middlewares,
	}
}

func (s *StaticMenu[User]) getReplyText() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyText
}

func (s *StaticMenu[User]) getReplyTextLanguageKey() string {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.replyWithLanguageKey
}

func (s *StaticMenu[User]) getDeferredReplyText() DeferredTextBuilder[User] {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.deferredReplyText
}

// processReplyText with StateUpdate[User] and returns the text to be replied.
func (s *StaticMenu[User]) processReplyText(update *StateUpdate[User]) string {
	if s.getDeferredReplyText() != nil {
		return s.getDeferredReplyText()(update)
	}

	if s.getReplyText() != "" {
		return s.getReplyText()
	}

	if s.getReplyTextLanguageKey() != "" {
		return update.Language().MustGet(s.getReplyTextLanguageKey())
	}

	return ""
}

func (s *StaticMenu[User]) processActionBuilder(client *tgbotapi.TelegramBot, update *StateUpdate[User]) *ActionBuilder {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.deferredActionBuilder != nil {
		return s.deferredActionBuilder(client, update)
	}

	return s.actionBuilder
}
