package app

import (
	"github.com/adimail/sisyphus/manager"
	"github.com/adimail/sisyphus/ui"

	"github.com/awesome-gocui/gocui"
)

type TodoApp struct {
	gui         *gocui.Gui
	taskManager *manager.TaskManager
}

func NewTodoApp() (*TodoApp, error) {
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, err
	}

	todoApp := &TodoApp{
		gui:         g,
		taskManager: manager.NewTaskManager(),
	}

	return todoApp, nil
}

func (app *TodoApp) Initialize() error {
	app.gui.SetManagerFunc(func(g *gocui.Gui) error {
		return ui.Layout(app.gui, app.taskManager)
	})

	if err := ui.ConfigureKeybindings(app.gui, app.taskManager); err != nil {
		return err
	}

	return nil
}

func (app *TodoApp) Run() error {
	defer app.gui.Close()

	if err := app.Initialize(); err != nil {
		return err
	}

	return app.gui.MainLoop()
}
