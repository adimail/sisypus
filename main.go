package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type Task struct {
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}

var tasks []Task

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorStrike = "\033[9m"
)

var selectedTaskIndex int // Tracks the currently selected task

func getTasksFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "sisyphus", "tasks.json"), nil
}

func loadTasks() error {
	tasksFile, err := getTasksFilePath()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	file, err := os.Open(tasksFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create an empty tasks.json if it doesn't exist
			if err := saveTasks(); err != nil {
				return fmt.Errorf("failed to create tasks file: %w", err)
			}
			return nil
		}
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.Size() == 0 {
		tasks = []Task{}
		return nil
	}

	return json.NewDecoder(file).Decode(&tasks)
}

func saveTasks() error {
	tasksFile, err := getTasksFilePath()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Ensure the directory exists before creating the file
	dir := filepath.Dir(tasksFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(tasksFile)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(tasks)
}

func main() {
	if err := loadTasks(); err != nil {
		log.Panicln("Failed to load tasks:", err)
	}
	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	// Quit keybindings
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// Navigation keybindings
	if err := g.SetKeybinding("list", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("list", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}

	// Add and toggle keybindings
	if err := g.SetKeybinding("", 'a', gocui.ModNone, addTask); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("list", gocui.KeySpace, gocui.ModNone, toggleTask); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, confirmClearTasks); err != nil {
		log.Panicln(err)
	}

	// Show Help Menu
	if err := g.SetKeybinding("", 'h', gocui.ModNone, showHelp); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("list", 'd', gocui.ModNone, confirmDeleteTask); err != nil {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Create the task list view
	if v, err := g.SetView("list", maxX/8, maxY/8, 7*maxX/8, 7*maxY/8, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = "Tasks"

		if _, err := g.SetCurrentView("list"); err != nil {
			return err
		}
	}

	// Update the status bar
	updateStatusBar(g)

	// Refresh the task list
	return refreshTaskList(g)
}

func refreshTaskList(g *gocui.Gui) error {
	v, err := g.View("list")
	if err != nil {
		return err
	}
	v.Clear()
	v.Wrap = true
	v.Autoscroll = true

	if len(tasks) == 0 {
		fmt.Fprintln(v, "Press 'a' to add a new task")
		return nil
	}

	for i, task := range tasks {
		indicator := "   "
		taskColor := colorGreen
		checkbox := "[ ]"

		if i == selectedTaskIndex {
			indicator = "-> "
		}

		if task.Completed {
			taskColor = colorRed
			task.Name = strikeThrough(task.Name)
			checkbox = "[X]"
		}

		fmt.Fprintf(v, "%s%s%s %s%s\n", indicator, taskColor, checkbox, task.Name, colorReset)
	}
	return nil
}

func updateStatusBar(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Define the position of the status bar at the bottom
	v, err := g.SetView("status", 0, maxY-2, maxX-1, maxY, 0)
	if err != nil {
		// Ignore the error if it's because the view is being created for the first time
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
	}

	v.Frame = false
	v.Clear()

	completedTasks := 0
	for _, task := range tasks {
		if task.Completed {
			completedTasks++
		}
	}
	totalTasks := len(tasks)
	percentageCompleted := 0
	if totalTasks > 0 {
		percentageCompleted = (completedTasks * 100) / totalTasks
	}

	fmt.Fprintf(v, "Sisyphus | %s%d%% Completed%s (%d/%d) | Help [h]",
		colorGreen, percentageCompleted, colorReset, completedTasks, totalTasks)

	return nil
}

func showHelp(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if helpView, err := g.SetView("help", maxX/4, maxY/4-5, 3*maxX/4, maxY/4+5, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		helpView.Title = "Help - Keybindings"
		helpView.Autoscroll = true
		helpView.Wrap = true

		// Set the background color to gray (ANSI code 48;5;235)
		helpView.BgColor = gocui.ColorBlack // This is required for 'gocui' compatibility
		helpView.FgColor = gocui.ColorWhite // Set foreground text color to white for contrast

		// Custom gray background using ANSI code
		helpView.Highlight = false // Disable text highlighting

		helpText := `Keybindings:
- Arrow Up: Move up the task list
- Arrow Down: Move down the task list
- Space: Toggle task completion
- 'a': Add a new task
- 'd': Delete selected task
- 'q' or Ctrl+C: Quit
- 'h': Show this help menu
- Ctrl+D: Clear all tasks`

		// Display the help text
		fmt.Fprintln(helpView, helpText)

		// Keybinding to close the help menu
		if err := g.SetKeybinding("help", 'q', gocui.ModNone, closeHelp); err != nil {
			return err
		}

		// Set the current view to the help view
		if _, err := g.SetCurrentView("help"); err != nil {
			return err
		}
	}
	return nil
}

func closeHelp(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView("help")
	g.SetCurrentView("list")
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if selectedTaskIndex > 0 {
		selectedTaskIndex--
		return refreshTaskList(g)
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if selectedTaskIndex < len(tasks)-1 {
		selectedTaskIndex++
		return refreshTaskList(g)
	}
	return nil
}

func toggleTask(g *gocui.Gui, v *gocui.View) error {
	if selectedTaskIndex < len(tasks) {
		tasks[selectedTaskIndex].Completed = !tasks[selectedTaskIndex].Completed
		if err := saveTasks(); err != nil {
			return fmt.Errorf("failed to save tasks: %w", err)
		}
	}
	return refreshTaskList(g)
}

func addTask(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if inputView, err := g.SetView("input", maxX/4, maxY/4-3, 3*maxX/4, maxY/4-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		inputView.Title = "Add Task"
		inputView.Editable = true
		if _, err := g.SetCurrentView("input"); err != nil {
			return err
		}
		if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, saveTask); err != nil {
			return err
		}
		if err := g.SetKeybinding("input", gocui.KeyCtrlQ, gocui.ModNone, cancelAddTask); err != nil {
			return err
		}
	}
	return nil
}

func saveTask(g *gocui.Gui, v *gocui.View) error {
	taskName := strings.TrimSpace(v.Buffer())
	if taskName != "" {
		tasks = append(tasks, Task{Name: taskName})
		if err := saveTasks(); err != nil {
			return fmt.Errorf("failed to save tasks: %w", err)
		}
	}
	g.DeleteView("input")
	g.SetCurrentView("list")
	return refreshTaskList(g)
}

func confirmClearTasks(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if dialogView, err := g.SetView("confirm", maxX/4, maxY/2-1, 3*maxX/4, maxY/2+1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		dialogView.Title = "Confirm"
		dialogView.Editable = false
		fmt.Fprintln(dialogView, "Clear all tasks? (y/n)")

		if err := g.SetKeybinding("confirm", 'y', gocui.ModNone, clearTasks); err != nil {
			return err
		}
		if err := g.SetKeybinding("confirm", 'n', gocui.ModNone, cancelClearTasks); err != nil {
			return err
		}
		if _, err := g.SetCurrentView("confirm"); err != nil {
			return err
		}
	}
	return nil
}

func clearTasks(g *gocui.Gui, v *gocui.View) error {
	tasks = []Task{}
	selectedTaskIndex = 0
	if err := saveTasks(); err != nil {
		return fmt.Errorf("failed to save tasks: %w", err)
	}
	g.DeleteView("confirm")
	g.SetCurrentView("list")
	return refreshTaskList(g)
}

func cancelClearTasks(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView("confirm")
	g.SetCurrentView("list")
	return nil
}

func cancelAddTask(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView("input")
	g.SetCurrentView("list")
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func strikeThrough(text string) string {
	return colorStrike + text + colorReset
}

func confirmDeleteTask(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if dialogView, err := g.SetView("confirmDelete", maxX/4, maxY/2-1, 3*maxX/4, maxY/2+1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		dialogView.Title = "Confirm Deletion"
		fmt.Fprintf(dialogView, "Delete task '%s'? (y/n)", tasks[selectedTaskIndex].Name)

		g.SetKeybinding("confirmDelete", 'y', gocui.ModNone, deleteTask)
		g.SetKeybinding("confirmDelete", 'n', gocui.ModNone, cancelDeleteTask)
		g.SetCurrentView("confirmDelete")
	}
	return nil
}

func deleteTask(g *gocui.Gui, v *gocui.View) error {
	tasks = append(tasks[:selectedTaskIndex], tasks[selectedTaskIndex+1:]...)
	selectedTaskIndex = 0
	g.DeleteView("confirmDelete")
	g.SetCurrentView("list")
	return refreshTaskList(g)
}

func cancelDeleteTask(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView("confirmDelete")
	g.SetCurrentView("list")
	return nil
}
