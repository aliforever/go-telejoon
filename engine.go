package telejoon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tgbotapi "github.com/aliforever/go-telegram-bot-api/v2"
)

// maxErrorMessageLen keeps error reports below Telegram's 4096-character message limit.
const maxErrorMessageLen = 4000

type engine struct {
	opts *Options
}

// newEngine returns an engine using the first non-nil Options, if any.
func newEngine(opts ...*Options) engine {
	for _, opt := range opts {
		if opt != nil {
			return engine{opts: opt}
		}
	}

	return engine{}
}

func (t *engine) onErr(
	ctx context.Context, client *tgbotapi.Bot, update tgbotapi.Update, err error) {

	j, _ := json.Marshal(update)

	msg := fmt.Sprintf("Error: %s\nUpdate: %s", err.Error(), string(j))
	if len(msg) > maxErrorMessageLen {
		msg = msg[:maxErrorMessageLen] + "... (truncated)"
	}

	if t.opts == nil {
		log.Print(msg)
		return
	}

	if t.opts.ErrorGroupID != 0 {
		_, sendErr := client.Message().
			ChatID(t.opts.ErrorGroupID).
			Text(msg).
			Send(ctx)
		if sendErr != nil {
			msg = fmt.Sprintf("%s\nerror_sending_to_error_group: %s", msg, sendErr.Error())
		}
	}

	if t.opts.Logger != nil {
		t.opts.Logger.Error(msg)
	} else {
		log.Print(msg)
	}
}
