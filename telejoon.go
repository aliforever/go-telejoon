package telejoon

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// Start long-polls for updates and blocks until the context is cancelled.
// This is the recommended way to start the bot as it supports graceful
// shutdown: on shutdown it waits for in-flight update handlers to finish
// before returning. In-flight handlers receive a context that is NOT
// cancelled, so their sends can complete.
//
// Updates are processed concurrently across chats, but updates that share a
// chat are processed one at a time, in arrival order: a user double-tapping
// a button or sending two quick messages can never race their own state
// transitions or duplicate side effects.
func Start(ctx context.Context, client *tgbotapi.Bot, processor Processor) error {
	return start(ctx, client, processor, 0)
}

// StartWithShutdownTimeout is Start, but graceful shutdown waits at most
// timeout for in-flight and queued updates to drain before returning.
// Processing goroutines are not killed; a handler outliving the timeout may
// still finish its sends in the background.
func StartWithShutdownTimeout(
	ctx context.Context,
	client *tgbotapi.Bot,
	processor Processor,
	timeout time.Duration,
) error {

	return start(ctx, client, processor, timeout)
}

func start(ctx context.Context, client *tgbotapi.Bot, processor Processor, shutdownTimeout time.Duration) error {
	var wg sync.WaitGroup

	serializer := &chatSerializer{tails: map[int64]chan struct{}{}}

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
			// Chain the update onto its chat's tail in the polling loop, so
			// the serialization order is the arrival order regardless of
			// goroutine scheduling.
			key, keyed := updateChatKey(update)

			var wait <-chan struct{}
			var link chan struct{}

			if keyed {
				wait, link = serializer.chain(key)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()

				if keyed {
					defer serializer.finish(key, link)

					if wait != nil {
						<-wait
					}
				}

				processor.Process(processCtx, client, update)
			}()
		}
	}

	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(done)
	}()

	if shutdownTimeout <= 0 {
		<-done

		return ctx.Err()
	}

	select {
	case <-done:
		return ctx.Err()
	case <-time.After(shutdownTimeout):
		return fmt.Errorf("telejoon: shutdown timed out after %s with updates still in flight", shutdownTimeout)
	}
}

// updateChatKey returns the serialization key of an update: its chat ID.
// Updates without a chat are not serialized.
func updateChatKey(update tgbotapi.Update) (int64, bool) {
	if chat := update.Chat(); chat != nil {
		return chat.Id, true
	}

	return 0, false
}

// chatSerializer serializes updates that share a chat using a chain of
// channels: each update waits for its predecessor's link to close. Chains
// are linked in the polling loop, so the serialization order is exactly the
// arrival order, and finished links drop out of the map.
type chatSerializer struct {
	mu    sync.Mutex
	tails map[int64]chan struct{}
}

// chain appends a link to the chat's chain and returns the channel to wait
// on (nil when the chain was empty) and the link to pass to finish.
func (s *chatSerializer) chain(key int64) (<-chan struct{}, chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.tails[key]
	link := make(chan struct{})
	s.tails[key] = link

	return prev, link
}

// finish releases the next update in the chain and unlinks a finished tail.
func (s *chatSerializer) finish(key int64, link chan struct{}) {
	close(link)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tails[key] == link {
		delete(s.tails, key)
	}
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
