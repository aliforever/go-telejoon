package telejoon

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/aliforever/go-telegram-bot-api"
	"log"
)

type engine[User any, Channel any, Group any, Language any] struct {
	opts []*Options
}

func (t *engine[User, Channel, Group, Language]) onErr(
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
