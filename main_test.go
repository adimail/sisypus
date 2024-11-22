package main

import (
	"os"
	"strings"
	"testing"
)

func TestGetTasksFilePath(t *testing.T) {
	path, err := getTasksFilePath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !strings.Contains(path, "sisyphus/tasks.json") {
		t.Errorf("Expected path to contain 'sisyphus/tasks.json', got %v", path)
	}
}

func TestSaveAndLoadTasks(t *testing.T) {
	tempDir := t.TempDir()
	originalPath := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalPath)

	testTasks := []Task{
		{Name: "Task 1", Completed: false},
		{Name: "Task 2", Completed: true},
	}

	tasks = testTasks
	err := saveTasks()
	if err != nil {
		t.Fatalf("Failed to save tasks: %v", err)
	}

	tasks = nil

	err = loadTasks()
	if err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	if len(tasks) != len(testTasks) {
		t.Fatalf("Expected %d tasks, got %d", len(testTasks), len(tasks))
	}
	for i, task := range testTasks {
		if tasks[i].Name != task.Name || tasks[i].Completed != task.Completed {
			t.Errorf("Task mismatch. Expected %v, got %v", task, tasks[i])
		}
	}
}

func TestAddTask(t *testing.T) {
	tasks = []Task{}
	newTask := Task{Name: "New Task", Completed: false}

	tasks = append(tasks, newTask)

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0] != newTask {
		t.Errorf("Expected task %v, got %v", newTask, tasks[0])
	}
}

func TestToggleTask(t *testing.T) {
	tasks = []Task{
		{Name: "Task 1", Completed: false},
	}

	selectedTaskIndex = 0
	tasks[selectedTaskIndex].Completed = !tasks[selectedTaskIndex].Completed

	if !tasks[0].Completed {
		t.Errorf("Expected task to be completed, got %v", tasks[0].Completed)
	}
}

func TestClearTasks(t *testing.T) {
	tasks = []Task{
		{Name: "Task 1", Completed: false},
		{Name: "Task 2", Completed: true},
	}

	tasks = []Task{}

	if len(tasks) != 0 {
		t.Fatalf("Expected 0 tasks, got %d", len(tasks))
	}
}
