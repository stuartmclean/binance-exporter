package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/tideland/golib/logger"
)

const (
	DefaultAction = "report"
)

func main() {
	exitIfErr(godotenv.Load(), "Error loading .env file")
	logger.SetLevelString(os.Getenv("LOG_LEVEL"))

	switch action() {
	case DefaultAction:
		NewApp().Report()
	case "run":
		NewApp().Run()
	case "h", "help":
		printMan()
	default:
		printMan()
	}
}

func action() string {
	action := DefaultAction
	if len(os.Args) > 1 {
		action = os.Args[1]
	}
	return action
}

func printMan() {
	man := "usage: binancebot <action>\n\treport (default)\n\trun\n"
	fmt.Print(man)
}
