package telejoon

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aliforever/go-telegram-bot-api"
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

// sendConfigWithErrHandler is a helper function to send a message with a config and handle errors.
func (t *engine) sendConfigWithErrHandler(
	client *tgbotapi.TelegramBot, config tgbotapi.Config, update tgbotapi.Update) (*tgbotapi.Response, error) {

	if t.opts != nil && t.opts.Logger != nil {
		j, _ := json.Marshal(config)
		t.opts.Logger.Infof("Sending message: %s", string(j))
	}

	resp, err := client.Send(config)
	if err != nil {
		t.onErr(client, update, err)
		return nil, err
	}

	return resp, nil
}

func (t *engine) onErr(
	client *tgbotapi.TelegramBot, update tgbotapi.Update, err error) {

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
		_, sendErr := client.Send(client.Message().
			SetChatId(t.opts.ErrorGroupID).
			SetText(msg))
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
