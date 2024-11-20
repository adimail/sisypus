package main

import (
	"errors"
	"log"

	"github.com/adimail/sisyphus/app"
	"github.com/awesome-gocui/gocui"
)

func main() {
	todoApp, err := app.NewTodoApp()
	if err != nil {
		log.Panicln(err)
	}

	if err := todoApp.Run(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
}
