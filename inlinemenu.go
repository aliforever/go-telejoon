package telejoon

import (
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"sync"
)

type InlineMenu[User any] struct {
	lock sync.Mutex

	middlewares []func(*tgbotapi.TelegramBot, *StateUpdate[User]) bool

	replyText         string
	deferredReplyText func(update *StateUpdate[User]) string

	callbackPrefix string

	inlineActionBuilder   *InlineActionBuilder
	deferredActionBuilder func(update *StateUpdate[User]) *InlineActionBuilder
}

func NewInlineMenuWithTextAndActionBuilder[User any](text string, builder *InlineActionBuilder) *InlineMenu[User] {
	return &InlineMenu[User]{
		replyText:           text,
		inlineActionBuilder: builder,
	}
}

func NewInlineMenuWithTextAndDeferredActionBuilder[User any](
	text string, deferredBuilder func(update *StateUpdate[User]) *InlineActionBuilder) *InlineMenu[User] {

	return &InlineMenu[User]{
		replyText:             text,
		deferredActionBuilder: deferredBuilder,
	}
}

func NewInlineMenuWithDeferredTextAndDeferredActionBuilder[User any](
	deferredText func(update *StateUpdate[User]) string,
	builder *InlineActionBuilder) *InlineMenu[User] {

	return &InlineMenu[User]{
		deferredReplyText:   deferredText,
		inlineActionBuilder: builder,
	}
}

func NewInlineMenuWithDeferredTextAndActionBuilder[User any](
	deferredText func(update *StateUpdate[User]) string,
	deferredBuilder func(update *StateUpdate[User]) *InlineActionBuilder) *InlineMenu[User] {

	return &InlineMenu[User]{
		deferredReplyText:     deferredText,
		deferredActionBuilder: deferredBuilder,
	}
}

// AddMiddleware adds a middleware to the inline menu
func (i *InlineMenu[User]) AddMiddleware(
	middleware func(*tgbotapi.TelegramBot, *StateUpdate[User]) bool) *InlineMenu[User] {

	i.lock.Lock()
	defer i.lock.Unlock()

	i.middlewares = append(i.middlewares, middleware)

	return i
}

// getMiddlewares returns the middlewares.
func (i *InlineMenu[User]) getMiddlewares() []func(bot *tgbotapi.TelegramBot, update *StateUpdate[User]) bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.middlewares
}
