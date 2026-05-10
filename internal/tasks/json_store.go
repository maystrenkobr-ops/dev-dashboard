package tasks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type JSONStore struct {
	file   string
	tasks  []Task
	nextID int
	mu     sync.Mutex
}

func NewJSONStore(file string) *JSONStore {
	store := &JSONStore{
		file:   file,
		tasks:  []Task{},
		nextID: 1,
	}

	store.load()

	return store
}

func (s *JSONStore) Close() {}

func (s *JSONStore) GetTasks(ctx context.Context) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := append([]Task(nil), s.tasks...)
	return result, nil
}

func (s *JSONStore) CreateTask(ctx context.Context, input Task) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := Task{
		ID:       s.nextID,
		Title:    input.Title,
		Status:   input.Status,
		Priority: input.Priority,
		Deadline: input.Deadline,
	}

	s.nextID++
	s.tasks = append(s.tasks, task)

	return task, s.saveLocked()
}

func (s *JSONStore) UpdateTaskTitle(ctx context.Context, id int, title string) (Task, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Title = title
			return s.tasks[i], true, s.saveLocked()
		}
	}

	return Task{}, false, nil
}

func (s *JSONStore) UpdateTaskDeadline(ctx context.Context, id int, deadline string) (Task, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Deadline = deadline
			return s.tasks[i], true, s.saveLocked()
		}
	}

	return Task{}, false, nil
}

func (s *JSONStore) UpdateTaskStatus(ctx context.Context, id int, status string) (Task, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Status = status
			return s.tasks[i], true, s.saveLocked()
		}
	}

	return Task{}, false, nil
}

func (s *JSONStore) UpdateTaskPriority(ctx context.Context, id int, priority string) (Task, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.tasks {
		if s.tasks[i].ID == id {
			s.tasks[i].Priority = priority
			return s.tasks[i], true, s.saveLocked()
		}
	}

	return Task{}, false, nil
}

func (s *JSONStore) DeleteTask(ctx context.Context, id int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, task := range s.tasks {
		if task.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return true, s.saveLocked()
		}
	}

	return false, nil
}

func (s *JSONStore) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		s.tasks = []Task{
			{ID: 1, Title: "Создать первый мини-проект", Status: "done", Priority: "medium", Deadline: ""},
			{ID: 2, Title: "Добавить PostgreSQL", Status: "done", Priority: "high", Deadline: ""},
			{ID: 3, Title: "Разнести backend-код по структуре", Status: "todo", Priority: "high", Deadline: ""},
		}
		s.nextID = 4
		_ = s.saveLocked()
		return
	}

	if err := json.Unmarshal(data, &s.tasks); err != nil {
		s.tasks = []Task{}
		s.nextID = 1
		return
	}

	changed := false
	s.nextID = 1

	for i := range s.tasks {
		if s.tasks[i].ID >= s.nextID {
			s.nextID = s.tasks[i].ID + 1
		}

		if s.tasks[i].Status == "" {
			s.tasks[i].Status = "todo"
			changed = true
		}

		if s.tasks[i].Priority == "" {
			s.tasks[i].Priority = "medium"
			changed = true
		}
	}

	if changed {
		_ = s.saveLocked()
	}
}

func (s *JSONStore) saveLocked() error {
	dir := filepath.Dir(s.file)

	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(s.tasks, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}
