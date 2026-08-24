package telejoon

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// GroupUpdateHandler is a function that handles group updates.
// Return true to continue processing with the next handler, false to stop.
type GroupUpdateHandler func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool

// GroupMiddleware runs before group handlers. Return false to stop processing.
type GroupMiddleware func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool

// EngineWithGroupHandlers handles group and supergroup chat updates
type EngineWithGroupHandlers struct {
	engine

	m            sync.RWMutex
	handlers     []GroupUpdateHandler
	middlewares  []GroupMiddleware
	panicHandler PanicHandler
}

// NewGroupHandlers creates a new group handler engine
func NewGroupHandlers(opts ...*Options) *EngineWithGroupHandlers {
	return &EngineWithGroupHandlers{
		engine: newEngine(opts...),
	}
}

// AddHandler adds a handler for group updates.
// Handlers are called in order until one returns false.
func (e *EngineWithGroupHandlers) AddHandler(handler GroupUpdateHandler) *EngineWithGroupHandlers {
	e.m.Lock()
	defer e.m.Unlock()

	e.handlers = append(e.handlers, handler)
	return e
}

// AddMiddleware adds a middleware for group updates.
// Middlewares are called before handlers. Return false to stop processing.
func (e *EngineWithGroupHandlers) AddMiddleware(middleware GroupMiddleware) *EngineWithGroupHandlers {
	e.m.Lock()
	defer e.m.Unlock()

	e.middlewares = append(e.middlewares, middleware)
	return e
}

// WithPanicHandler sets the panic handler
func (e *EngineWithGroupHandlers) WithPanicHandler(handler PanicHandler) *EngineWithGroupHandlers {
	e.m.Lock()
	defer e.m.Unlock()

	e.panicHandler = handler
	return e
}

func (e *EngineWithGroupHandlers) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil {
		return chat.Type == "group" || chat.Type == "supergroup"
	}
	return false
}

func (e *EngineWithGroupHandlers) Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	e.m.RLock()
	middlewares := append([]GroupMiddleware(nil), e.middlewares...)
	handlers := append([]GroupUpdateHandler(nil), e.handlers...)
	panicHandler := e.panicHandler
	e.m.RUnlock()

	// Always recover: each update runs in its own goroutine, so an
	// unrecovered panic would take down the whole process.
	defer func() {
		if r := recover(); r != nil {
			if panicHandler != nil {
				panicHandler(client, update, r, string(debug.Stack()))
			} else {
				e.onErr(ctx, client, update, fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
			}
		}
	}()

	// Run middlewares
	for _, mw := range middlewares {
		if !mw(ctx, client, update) {
			return
		}
	}

	// Run handlers
	for _, handler := range handlers {
		if !handler(ctx, client, update) {
			return
		}
	}
}

// ChannelUpdateHandler is a function that handles channel updates.
// Return true to continue processing with the next handler, false to stop.
type ChannelUpdateHandler func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool

// ChannelMiddleware runs before channel handlers. Return false to stop processing.
type ChannelMiddleware func(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) bool

// EngineWithChannelHandlers handles channel updates
type EngineWithChannelHandlers struct {
	engine

	m            sync.RWMutex
	handlers     []ChannelUpdateHandler
	middlewares  []ChannelMiddleware
	panicHandler PanicHandler
}

// NewChannelHandlers creates a new channel handler engine
func NewChannelHandlers(opts ...*Options) *EngineWithChannelHandlers {
	return &EngineWithChannelHandlers{
		engine: newEngine(opts...),
	}
}

// AddHandler adds a handler for channel updates.
// Handlers are called in order until one returns false.
func (e *EngineWithChannelHandlers) AddHandler(handler ChannelUpdateHandler) *EngineWithChannelHandlers {
	e.m.Lock()
	defer e.m.Unlock()

	e.handlers = append(e.handlers, handler)
	return e
}

// AddMiddleware adds a middleware for channel updates.
// Middlewares are called before handlers. Return false to stop processing.
func (e *EngineWithChannelHandlers) AddMiddleware(middleware ChannelMiddleware) *EngineWithChannelHandlers {
	e.m.Lock()
	defer e.m.Unlock()

	e.middlewares = append(e.middlewares, middleware)
	return e
}

// WithPanicHandler sets the panic handler
func (e *EngineWithChannelHandlers) WithPanicHandler(handler PanicHandler) *EngineWithChannelHandlers {
	e.m.Lock()
	defer e.m.Unlock()

	e.panicHandler = handler
	return e
}

func (e *EngineWithChannelHandlers) canProcess(update tgbotapi.Update) bool {
	if chat := update.Chat(); chat != nil {
		return chat.Type == "channel"
	}
	// Also handle channel posts (updates without Chat when it's a channel post)
	if update.ChannelPost != nil {
		return true
	}
	return false
}

func (e *EngineWithChannelHandlers) Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update) {
	e.m.RLock()
	middlewares := append([]ChannelMiddleware(nil), e.middlewares...)
	handlers := append([]ChannelUpdateHandler(nil), e.handlers...)
	panicHandler := e.panicHandler
	e.m.RUnlock()

	// Always recover: each update runs in its own goroutine, so an
	// unrecovered panic would take down the whole process.
	defer func() {
		if r := recover(); r != nil {
			if panicHandler != nil {
				panicHandler(client, update, r, string(debug.Stack()))
			} else {
				e.onErr(ctx, client, update, fmt.Errorf("panic: %v\n%s", r, debug.Stack()))
			}
		}
	}()

	// Run middlewares
	for _, mw := range middlewares {
		if !mw(ctx, client, update) {
			return
		}
	}

	// Run handlers
	for _, handler := range handlers {
		if !handler(ctx, client, update) {
			return
		}
	}
}
