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
		dt         bool
		dah        bool
	)

	flag.StringVar(&token, "token", "", "Bot token")
	flag.StringVar(&modulePath, "module_path", "", "Module Path")
	flag.BoolVar(&dt, "dt", false, "Print Deferred Text Function")
	flag.BoolVar(&dah, "dah", false, "Print Deferred Action Handler Function")

	flag.Parse()

	if token != "" && modulePath != "" {
		err := cmd.NewGenerator(token, modulePath).Generate()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	if dt {
		cmd.NewPrinter().PrintDeferredTextFunction()
		return
	}

	if dah {
		cmd.NewPrinter().PrintDeferredActionHandlerFunction()
		return
	}

	fmt.Println("Choose one of the following commands:")
	fmt.Println("1. Generate new project: tgbot -token <token> -module_path <module_path>")
	fmt.Println("2. Print Deferred Text Function tgbot -dt")
	fmt.Println("3. Print Deferred Action Handler Function tgbot -dah")

	fmt.Println("Enter number: ")

	var choice int

	_, err := fmt.Scanln(&choice)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch choice {
	case 1:
		fmt.Println("Enter token: ")
		_, err = fmt.Scanln(&token)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Enter module path: ")
		_, err = fmt.Scanln(&modulePath)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = cmd.NewGenerator(token, modulePath).Generate()
		if err != nil {
			fmt.Println(err)
			return
		}
	case 2:
		cmd.NewPrinter().PrintDeferredTextFunction()
	case 3:
		cmd.NewPrinter().PrintDeferredActionHandlerFunction()
	default:
		fmt.Println("Invalid choice")
	}
}
