package ui

import (
	"errors"
	"strings"

	"github.com/adimail/sisyphus/manager"
	"github.com/awesome-gocui/gocui"
)

func ConfigureKeybindings(gui *gocui.Gui, taskManager *manager.TaskManager) error {
	// Quit binding
	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quitApp); err != nil {
		return err
	}

	// Movement: navigating through the tasks
	if err := gui.SetKeybinding("list", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := gui.SetKeybinding("list", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	// Movement: navigating between sections (Daily, Weekly, Monthly)
	if err := gui.SetKeybinding("list", gocui.KeyArrowLeft, gocui.ModNone, switchSection(gui, taskManager, "daily")); err != nil {
		return err
	}
	if err := gui.SetKeybinding("list", gocui.KeyArrowRight, gocui.ModNone, switchSection(gui, taskManager, "weekly")); err != nil {
		return err
	}
	if err := gui.SetKeybinding("list", 'l', gocui.ModNone, switchSection(gui, taskManager, "monthly")); err != nil {
		return err
	}

	// Task actions
	if err := gui.SetKeybinding("", 'a', gocui.ModNone, addTask(gui, taskManager)); err != nil {
		return err
	}
	if err := gui.SetKeybinding("list", gocui.KeySpace, gocui.ModNone, toggleTask(gui, taskManager)); err != nil {
		return err
	}

	return nil
}

// Switch between sections (Daily, Weekly, Monthly)
func switchSection(gui *gocui.Gui, taskManager *manager.TaskManager, section string) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		taskManager.SwitchList(section)
		return RefreshTaskList(g, taskManager)
	}
}

// Quit the application
func quitApp(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// Navigate up in the tasks list
func cursorUp(g *gocui.Gui, v *gocui.View) error {
	_, oy := v.Origin()
	_, cy := v.Cursor()

	// Move cursor up or scroll up
	if cy > 0 {
		return v.SetCursor(0, cy-1)
	} else if oy > 0 {
		return v.SetOrigin(0, oy-1)
	}
	return nil
}

// Navigate down in the tasks list
func cursorDown(g *gocui.Gui, v *gocui.View) error {
	_, oy := v.Origin()
	_, cy := v.Cursor()
	tasks := v.BufferLines()

	// Move cursor down or scroll down
	if cy+1 < len(tasks) {
		_, height := v.Size()
		if cy+1 < height {
			return v.SetCursor(0, cy+1)
		} else {
			return v.SetOrigin(0, oy+1)
		}
	}
	return nil
}

// Toggle the task's completion state
func toggleTask(gui *gocui.Gui, taskManager *manager.TaskManager) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		_, cy := v.Cursor()
		taskManager.ToggleTask(cy)
		return RefreshTaskList(g, taskManager)
	}
}

// Add a task to the current section
func addTask(gui *gocui.Gui, taskManager *manager.TaskManager) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		maxX, maxY := g.Size()
		if inputView, err := g.SetView("input", maxX/4, maxY/4-3, 3*maxX/4, maxY/4-1, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			inputView.Title = "Add Task"
			inputView.Editable = true

			// Set focus on input view
			g.SetCurrentView("input")
			g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, inputView *gocui.View) error {
				taskName := strings.TrimSpace(inputView.Buffer())
				if taskName != "" {
					taskManager.AddTaskToCurrentList(taskName)
				}
				gui.DeleteView("input")
				gui.SetCurrentView("list")
				return RefreshTaskList(gui, taskManager)
			})
		}
		return nil
	}
}
