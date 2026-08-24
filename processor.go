package telejoon

import (
	"context"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// Processor is anything that can accept updates from the polling loop.
//
// Use Start to drive it: Start is the framework's supported polling loop and
// the layer that guarantees per-chat serialization — updates sharing a chat
// are processed one at a time, in arrival order. Engine.Process is safe for
// concurrent use, but if you hand-roll a loop over client.Updates instead of
// Start, ordering is yours to enforce: the same user's updates may otherwise
// interleave and race on state transitions.
type Processor interface {
	canProcess(update tgbotapi.Update) bool
	Process(ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update)
}
