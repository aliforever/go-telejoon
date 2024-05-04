# go-telejoon
Telegram bot framework using generics making it easier to write bots.

## TODOs
- [ ] Add more handlers for groups, channels, etc. (currently identified as middleware)
- [ ] Change TextBuilder for language to be identified using `{{Title}}` instead of LanguageTextBuilder
## Note:
- This is a work in progress and is not ready for production use.

## Docs
[Here](https://pkg.go.dev/github.com/aliforever/go-telejoon)

## Install Project Generator:
```bash
go install github.com/aliforever/go-telejoon/tgbot
```

## Generate Project:
```bash
tgbot --token=BOT_TOKEN_HERE --module_path=MODULE_PATH_HERE
```
