package manager

type Task struct {
	Name      string
	Completed bool
}

type TaskManager struct {
	DailyGoals   []Task
	WeeklyGoals  []Task
	MonthlyGoals []Task
	CurrentList  *[]Task
}

func NewTaskManager() *TaskManager {
	tm := &TaskManager{
		DailyGoals: []Task{
			{"Improve UI of Sisyphus", false},
			{"Fix arrrow key navigation bugs", false},
			{"add status bar at bottom", false},
		},
		WeeklyGoals: []Task{
			{"Weekly Task 1", false},
			{"Weekly Task 2", false},
		},
		MonthlyGoals: []Task{
			{"Monthly Task 1", false},
			{"Monthly Task 2", false},
		},
	}
	tm.CurrentList = &tm.DailyGoals // Default to DailyGoals section
	return tm
}

func (tm *TaskManager) AddTaskToCurrentList(name string) {
	*tm.CurrentList = append(*tm.CurrentList, Task{Name: name})
}

func (tm *TaskManager) ToggleTask(index int) {
	if index >= 0 && index < len(*tm.CurrentList) {
		(*tm.CurrentList)[index].Completed = !(*tm.CurrentList)[index].Completed
	}
}

func (tm *TaskManager) SwitchList(section string) {
	switch section {
	case "daily":
		tm.CurrentList = &tm.DailyGoals
	case "weekly":
		tm.CurrentList = &tm.WeeklyGoals
	case "monthly":
		tm.CurrentList = &tm.MonthlyGoals
	}
}

func (tm *TaskManager) GetCurrentTasks() []Task {
	return *tm.CurrentList
}
