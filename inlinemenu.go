package telejoon

import (
	"sync"
)

type InlineMenu struct {
	lock sync.Mutex

	middlewares []Middleware

	textBuilder TextBuilder

	callbackPrefix string

	inlineActionBuilder InlineActionBuilderKind
}

func NewInlineMenu(
	textBuilder TextBuilder,
	actionBuilder InlineActionBuilderKind,
	middlewares ...Middleware,
) *InlineMenu {

	return &InlineMenu{
		textBuilder:         textBuilder,
		inlineActionBuilder: actionBuilder,
		middlewares:         middlewares,
	}
}

// getMiddlewares returns the middlewares.
func (i *InlineMenu) getMiddlewares() []Middleware {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.middlewares
}

func (i *InlineMenu) processActionBuilder(update *StateUpdate) *InlineActionBuilder {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.inlineActionBuilder == nil {
		return nil
	}

	builder := i.inlineActionBuilder.Build(update)
	builder.inlineMenu = i.callbackPrefix

	return builder
}

func (i *InlineMenu) processTextBuilder(update *StateUpdate) string {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.textBuilder == nil {
		return ""
	}

	return i.textBuilder.String(update)
}
