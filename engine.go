package telejoon

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"log"
)

type engine struct {
	opts []*Options
}

// sendConfigWithErrHandler is a helper function to send a message with a config and handle errors.
func (t *engine) sendConfigWithErrHandler(
	client *tgbotapi.TelegramBot, config tgbotapi.Config, update tgbotapi.Update) (*tgbotapi.Response, error) {

	if len(t.opts) > 0 {
		if t.opts[0].Logger != nil {
			j, _ := json.Marshal(config)
			t.opts[0].Logger.Infof("Sending message: %s", string(j))
		}
	}

	_, err := client.Send(config)
	if err != nil {
		t.onErr(client, update, err)
		return nil, err
	}

	return nil, err
}

func (t *engine) onErr(
	client *tgbotapi.TelegramBot, update tgbotapi.Update, err error) {

	if len(t.opts) > 0 {
		j, _ := json.Marshal(update)
		if t.opts[0].ErrorGroupID == 0 && t.opts[0].Logger == nil {
			log.Printf("Error: %s\nUpdate: %s", err.Error(), string(j))
			return
		}

		if t.opts[0].ErrorGroupID != 0 {
			_, sendErr := client.Send(client.Message().
				SetChatId(t.opts[0].ErrorGroupID).
				SetText(fmt.Sprintf("Error: %s\nUpdate: %s", err.Error(), string(j))))
			if sendErr != nil {
				err = fmt.Errorf("%s\n%s", err.Error(), sendErr.Error())
			}
		}

		if t.opts[0].Logger != nil {
			t.opts[0].Logger.Errorf("Error: %s\nUpdate: %s", err.Error(), string(j))
		} else {
			log.Printf("Error: %s\nUpdate: %s", err.Error(), string(j))
		}
	}
}
