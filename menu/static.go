// Package menu provides fluent builders for creating static and inline menus.
package menu

import (
	"github.com/aliforever/go-telejoon"
	"github.com/aliforever/go-telejoon/buttons"
	"github.com/aliforever/go-telejoon/text"
)

// StaticBuilder builds a StaticMenu with a fluent API.
type StaticBuilder struct {
	message    text.Text
	buttons    *buttons.Builder
	handlers   map[string]telejoon.UpdateHandler
	middleware []telejoon.UpdateHandler
}

// Static creates a new static menu builder.
//
// Example:
//
//	menu.Static(text.L("Welcome.Message")).
//	    Buttons(
//	        buttons.GoTo(text.L("Nav.Home"), "Welcome"),
//	    ).
//	    OnText(handleText).
//	    Build()
func Static(message text.Text) *StaticBuilder {
	return &StaticBuilder{
		message:  message,
		handlers: make(map[string]telejoon.UpdateHandler),
	}
}

// Buttons sets the button collection for this menu.
func (b *StaticBuilder) Buttons(btns ...*buttons.Button) *StaticBuilder {
	b.buttons = buttons.Build(btns...)
	return b
}

// ButtonsBuilder sets buttons from an existing builder.
func (b *StaticBuilder) ButtonsBuilder(builder *buttons.Builder) *StaticBuilder {
	b.buttons = builder
	return b
}

// Middleware adds a middleware handler.
func (b *StaticBuilder) Middleware(fn telejoon.UpdateHandler) *StaticBuilder {
	b.middleware = append(b.middleware, fn)
	return b
}

// OnText handles text messages not matching any button.
func (b *StaticBuilder) OnText(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.TextHandler] = fn
	return b
}

// OnPhoto handles photo messages.
func (b *StaticBuilder) OnPhoto(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.PhotoHandler] = fn
	return b
}

// OnDocument handles document messages.
func (b *StaticBuilder) OnDocument(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.DocumentHandler] = fn
	return b
}

// OnVideo handles video messages.
func (b *StaticBuilder) OnVideo(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.VideoHandler] = fn
	return b
}

// OnVoice handles voice messages.
func (b *StaticBuilder) OnVoice(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.VoiceHandler] = fn
	return b
}

// OnAudio handles audio messages.
func (b *StaticBuilder) OnAudio(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.AudioHandler] = fn
	return b
}

// OnLocation handles location messages.
func (b *StaticBuilder) OnLocation(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.LocationHandler] = fn
	return b
}

// OnContact handles contact messages.
func (b *StaticBuilder) OnContact(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.ContactHandler] = fn
	return b
}

// OnSticker handles sticker messages.
func (b *StaticBuilder) OnSticker(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.StickerHandler] = fn
	return b
}

// OnAny handles any message type not handled by specific handlers.
func (b *StaticBuilder) OnAny(fn telejoon.UpdateHandler) *StaticBuilder {
	b.handlers[telejoon.DefaultHandler] = fn
	return b
}

// Build creates the StaticMenu.
func (b *StaticBuilder) Build() *telejoon.StaticMenu {
	var handlersSlice []telejoon.Handler

	// Add middlewares
	for _, mw := range b.middleware {
		handlersSlice = append(handlersSlice, telejoon.NewMiddleware(mw))
	}

	// Add dynamic handlers
	for handlerType, fn := range b.handlers {
		switch handlerType {
		case telejoon.TextHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerText(fn))
		case telejoon.PhotoHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerPhoto(fn))
		case telejoon.DocumentHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerDocument(fn))
		case telejoon.VideoHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerVideo(fn))
		case telejoon.VoiceHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerVoice(fn))
		case telejoon.AudioHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerAudio(fn))
		case telejoon.LocationHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerLocation(fn))
		case telejoon.ContactHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerContact(fn))
		case telejoon.StickerHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDynamicHandlerSticker(fn))
		case telejoon.DefaultHandler:
			handlersSlice = append(handlersSlice, telejoon.NewDefaultHandler(fn))
		}
	}

	var actionBuilder telejoon.ActionBuilderKind
	if b.buttons != nil {
		actionBuilder = b.buttons.Build()
	}

	return telejoon.NewStaticMenu(
		b.message.Builder(),
		actionBuilder,
		handlersSlice...,
	)
}
