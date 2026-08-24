package telejoon

import (
	"context"
	"log"
	"sync"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// Start long-polls for updates and blocks until the context is cancelled.
// This is the recommended way to start the bot as it supports graceful
// shutdown: on shutdown it waits for in-flight update handlers to finish
// before returning. In-flight handlers receive a context that is NOT
// cancelled, so their sends can complete.
func Start(ctx context.Context, client *tgbotapi.Bot, processor Processor) error {
	var wg sync.WaitGroup

	// In-flight updates outlive the polling context: WithoutCancel lets their
	// handlers finish (and send) while Start winds down.
	processCtx := context.WithoutCancel(ctx)

	for update, err := range client.Updates(ctx) {
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			// Transient polling error: Updates already backs off; keep going.
			log.Printf("telejoon: polling error: %v", err)

			continue
		}

		if processor.canProcess(update) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				processor.Process(processCtx, client, update)
			}()
		}
	}

	wg.Wait()

	return ctx.Err()
}

// StartWithCallback starts the bot with a callback that is called on shutdown.
// The callback is always called, whether the shutdown is due to context cancellation or channel closure.
func StartWithCallback(ctx context.Context, client *tgbotapi.Bot, processor Processor, onShutdown func()) error {
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
func StartWithoutContext(client *tgbotapi.Bot, processor Processor) {
	_ = Start(context.Background(), client, processor)
}
