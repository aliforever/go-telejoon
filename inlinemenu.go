package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type InlineMenu struct {
	lock sync.Mutex

	middlewares []InlineMiddleware

	replyText         string
	deferredReplyText func(update *StateUpdate) string

	callbackPrefix string

	inlineActionBuilder   *InlineActionBuilder
	deferredActionBuilder func(update *StateUpdate) *InlineActionBuilder
}

type InlineMiddleware func(*tgbotapi.TelegramBot, *StateUpdate) bool

func NewInlineMenuWithTextAndActionBuilder(
	text string,
	builder *InlineActionBuilder,
	middlewares ...InlineMiddleware,
) *InlineMenu {

	return &InlineMenu{
		replyText:           text,
		inlineActionBuilder: builder,
		middlewares:         middlewares,
	}
}

func NewInlineMenuWithTextAndDeferredActionBuilder(
	text string,
	deferredBuilder func(update *StateUpdate) *InlineActionBuilder,
	middlewares ...InlineMiddleware,
) *InlineMenu {

	return &InlineMenu{
		replyText:             text,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

func NewInlineMenuWithDeferredTextAndDeferredActionBuilder(
	deferredText func(update *StateUpdate) string,
	deferredBuilder func(update *StateUpdate) *InlineActionBuilder,
	middlewares ...InlineMiddleware,
) *InlineMenu {

	return &InlineMenu{
		deferredReplyText:     deferredText,
		deferredActionBuilder: deferredBuilder,
		middlewares:           middlewares,
	}
}

func NewInlineMenuWithDeferredTextAndActionBuilder(
	deferredText func(update *StateUpdate) string,
	builder *InlineActionBuilder,
	middlewares ...InlineMiddleware,
) *InlineMenu {

	return &InlineMenu{
		deferredReplyText:   deferredText,
		inlineActionBuilder: builder,
		middlewares:         middlewares,
	}
}

// getMiddlewares returns the middlewares.
func (i *InlineMenu) getMiddlewares() []InlineMiddleware {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.middlewares
}

func (i *InlineMenu) getActionBuilder() *InlineActionBuilder {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.inlineActionBuilder == nil {
		return nil
	}

	i.inlineActionBuilder.inlineMenu = i.callbackPrefix

	return i.inlineActionBuilder
}

func (i *InlineMenu) getDeferredActionBuilder(update *StateUpdate) *InlineActionBuilder {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.deferredActionBuilder == nil {
		return nil
	}

	builder := i.deferredActionBuilder(update)
	builder.inlineMenu = i.callbackPrefix

	return builder
}
