package telejoon

import (
	"context"
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api"
)

// Start starts the bot and blocks until the context is cancelled or the updates channel is closed.
// This is the recommended way to start the bot as it supports graceful shutdown:
// on shutdown it waits for in-flight update handlers to finish before returning.
func Start(ctx context.Context, client *tgbotapi.TelegramBot, processor Processor) error {
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return ctx.Err()
		case update, ok := <-client.Updates():
			if !ok {
				wg.Wait()
				return nil
			}
			if processor.canProcess(update) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					processor.Process(client, update)
				}()
			}
		}
	}
}

// StartWithCallback starts the bot with a callback that is called on shutdown.
// The callback is always called, whether the shutdown is due to context cancellation or channel closure.
func StartWithCallback(ctx context.Context, client *tgbotapi.TelegramBot, processor Processor, onShutdown func()) error {
	defer func() {
		if onShutdown != nil {
			onShutdown()
		}
	}()
	return Start(ctx, client, processor)
}

// StartWithoutContext starts the bot without context support.
// Deprecated: Use Start with context.Background() instead for new code.
// This function is kept for backward compatibility.
func StartWithoutContext(client *tgbotapi.TelegramBot, processor Processor) {
	for update := range client.Updates() {
		if processor.canProcess(update) {
			go processor.Process(client, update)
		}
	}
}
