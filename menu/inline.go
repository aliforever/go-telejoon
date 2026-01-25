package menu

import (
	"github.com/aliforever/go-telejoon"
	"github.com/aliforever/go-telejoon/inline"
	"github.com/aliforever/go-telejoon/text"
)

// InlineBuilder builds an InlineMenu with a fluent API.
type InlineBuilder struct {
	message    text.Text
	buttons    *inline.Builder
	middleware []telejoon.UpdateHandler
}

// Inline creates a new inline menu builder.
//
// Example:
//
//	menu.Inline(text.S("Choose an option")).
//	    Buttons(
//	        inline.URL(text.S("Website"), "https://..."),
//	        inline.Callback(text.S("Delete"), handler),
//	    ).
//	    Build()
func Inline(message text.Text) *InlineBuilder {
	return &InlineBuilder{
		message: message,
	}
}

// Buttons sets the inline button collection for this menu.
func (b *InlineBuilder) Buttons(btns ...*inline.Button) *InlineBuilder {
	b.buttons = inline.Build(btns...)
	return b
}

// ButtonsBuilder sets buttons from an existing builder.
func (b *InlineBuilder) ButtonsBuilder(builder *inline.Builder) *InlineBuilder {
	b.buttons = builder
	return b
}

// Middleware adds a middleware handler.
func (b *InlineBuilder) Middleware(fn telejoon.UpdateHandler) *InlineBuilder {
	b.middleware = append(b.middleware, fn)
	return b
}

// Build creates the InlineMenu.
func (b *InlineBuilder) Build() *telejoon.InlineMenu {
	var middlewares []telejoon.Middleware

	for _, mw := range b.middleware {
		middlewares = append(middlewares, telejoon.NewMiddleware(mw))
	}

	var actionBuilder telejoon.InlineActionBuilderKind
	if b.buttons != nil {
		actionBuilder = b.buttons.Build()
	}

	return telejoon.NewInlineMenu(
		b.message.Builder(),
		actionBuilder,
		middlewares...,
	)
}
