package tasks

import (
	"context"
	"time"
)

type Store interface {
	Close()
	GetTasks(ctx context.Context) ([]Task, error)
	CreateTask(ctx context.Context, input Task) (Task, error)
	UpdateTaskTitle(ctx context.Context, id int, title string) (Task, bool, error)
	UpdateTaskDeadline(ctx context.Context, id int, deadline string) (Task, bool, error)
	UpdateTaskStatus(ctx context.Context, id int, status string) (Task, bool, error)
	UpdateTaskPriority(ctx context.Context, id int, priority string) (Task, bool, error)
	DeleteTask(ctx context.Context, id int) (bool, error)
}

func NewStorage(ctx context.Context, databaseURL string, tasksFile string) (Store, error) {
	if databaseURL != "" {
		return NewPostgresStore(ctx, databaseURL)
	}

	return NewJSONStore(tasksFile), nil
}

func IsValidStatus(status string) bool {
	return status == "todo" || status == "in_progress" || status == "done"
}

func IsValidPriority(priority string) bool {
	return priority == "low" || priority == "medium" || priority == "high"
}

func IsValidDeadline(deadline string) bool {
	if deadline == "" {
		return true
	}

	_, err := time.Parse("2006-01-02", deadline)
	return err == nil
}
