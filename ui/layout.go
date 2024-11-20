package ui

import (
	"errors"
	"fmt"

	"github.com/adimail/sisyphus/manager"

	"github.com/awesome-gocui/gocui"
)

func Layout(gui *gocui.Gui, taskManager *manager.TaskManager) error {
	maxX, maxY := gui.Size()

	sections := []struct {
		Name   string
		StartX int
	}{
		{"Daily Goals", maxX / 12},
		{"Weekly Goals", 4 * maxX / 12},
		{"Monthly Goals", 7 * maxX / 12},
	}

	for _, sec := range sections {
		if v, err := gui.SetView(sec.Name, sec.StartX, maxY/6, sec.StartX+3*maxX/12-1, 5*maxY/6, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Title = sec.Name
			v.Highlight = true
			v.SelBgColor = gocui.ColorGreen
			v.SelFgColor = gocui.ColorBlack

			// Set the initial view focus to the first section (Daily Goals)
			if sec.Name == "Daily Goals" {
				gui.SetCurrentView(sec.Name)
			}
		}
	}

	// Status bar
	if v, err := gui.SetView("status", 0, maxY-2, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		v.BgColor = gocui.ColorBlue
		v.FgColor = gocui.ColorWhite
	}

	return RefreshTaskList(gui, taskManager)
}

func RefreshTaskList(gui *gocui.Gui, taskManager *manager.TaskManager) error {
	view := gui.CurrentView() // Ignore the error if you want, or handle it separately

	viewName := view.Name()
	switch viewName {
	case "Daily Goals":
		taskManager.SwitchList("daily")
	case "Weekly Goals":
		taskManager.SwitchList("weekly")
	case "Monthly Goals":
		taskManager.SwitchList("monthly")
	}

	view.Clear()
	for _, task := range taskManager.GetCurrentTasks() {
		if task.Completed {
			fmt.Fprintln(view, "\x1b[32m[X]\x1b[0m "+task.Name)
		} else {
			fmt.Fprintln(view, "[ ] "+task.Name)
		}
	}
	return nil
}
