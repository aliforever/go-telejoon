package main

import (
	"flag"
	"fmt"
	"github.com/aliforever/go-telejoon/tgbot/cmd"
)

func main() {
	var (
		token      string
		modulePath string
	)

	flag.StringVar(&token, "token", "", "Bot token")
	flag.StringVar(&modulePath, "module_path", "", "Module Path")

	flag.Parse()

	if token != "" && modulePath != "" {
		err := cmd.NewGenerator(token, modulePath).Generate()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
