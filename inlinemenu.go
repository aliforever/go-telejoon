package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type InlineMenu[User any] struct {
	lock sync.Mutex

	middlewares []InlineMiddleware[User]

	replyText         string
	deferredReplyText func(update *StateUpdate[User]) string

	callbackPrefix string

	inlineActionBuilder   *InlineActionBuilder
	deferredActionBuilder func(update *StateUpdate[User]) *InlineActionBuilder
}

type InlineMiddleware[User any] func(*tgbotapi.TelegramBot, *StateUpdate[User]) bool

func NewInlineMenuWithTextAndActionBuilder[User any](
	text string,
	builder *InlineActionBuilder,
	middlewares ...InlineMiddleware[User],
) *InlineMenu[User] {

	return &InlineMenu[User]{
		replyText:           text,
		inlineActionBuilder: builder,
		middlewares:         middlewares,
	}
}

func NewInlineMenuWithTextAndDeferredActionBuilder[User any](
	text string,
	deferredBuilder func(update *StateUpdate[User]) *InlineActionBuilder,
	middlewares ...InlineMiddleware[User],
) *InlineMenu[User] {

	return &InlineMenu[User]{
		replyText:             text,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

func NewInlineMenuWithDeferredTextAndDeferredActionBuilder[User any](
	deferredText func(update *StateUpdate[User]) string,
	deferredBuilder func(update *StateUpdate[User]) *InlineActionBuilder,
	middlewares ...InlineMiddleware[User],
) *InlineMenu[User] {

	return &InlineMenu[User]{
		deferredReplyText:     deferredText,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

func NewInlineMenuWithDeferredTextAndActionBuilder[User any](
	deferredText func(update *StateUpdate[User]) string,
	builder *InlineActionBuilder,
	middlewares ...InlineMiddleware[User],
) *InlineMenu[User] {

	return &InlineMenu[User]{
		deferredReplyText:   deferredText,
		inlineActionBuilder: builder,
		middlewares:         middlewares,
	}
}

// getMiddlewares returns the middlewares.
func (i *InlineMenu[User]) getMiddlewares() []InlineMiddleware[User] {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.middlewares
}

func (i *InlineMenu[User]) getActionBuilder() *InlineActionBuilder {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.inlineActionBuilder == nil {
		return nil
	}

	i.inlineActionBuilder.inlineMenu = i.callbackPrefix

	return i.inlineActionBuilder
}

func (i *InlineMenu[User]) getDeferredActionBuilder(update *StateUpdate[User]) *InlineActionBuilder {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.deferredActionBuilder == nil {
		return nil
	}

	builder := i.deferredActionBuilder(update)
	builder.inlineMenu = i.callbackPrefix

	return builder
}
