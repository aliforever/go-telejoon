package telejoon

import tgbotapi "github.com/aliforever/go-telegram-bot-api"

type Options struct {
	onErr func(update tgbotapi.Update, err error)
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) OnErr(onErr func(update tgbotapi.Update, err error)) *Options {
	o.onErr = onErr
	return o
}
