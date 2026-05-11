package tasks

import (
"context"
"encoding/json"
"os"
"path/filepath"
"sync"
"time"
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

func (s *JSONStore) EnsureWorkspaceSupport(ctx context.Context, defaultWorkspaceID int, defaultUserID int) error {
s.mu.Lock()
defer s.mu.Unlock()

changed := false

for i := range s.tasks {
if s.tasks[i].WorkspaceID == 0 {
s.tasks[i].WorkspaceID = defaultWorkspaceID
changed = true
}

if s.tasks[i].CreatedBy == 0 {
s.tasks[i].CreatedBy = defaultUserID
changed = true
}
}

if changed {
return s.saveLocked()
}

return nil
}

func (s *JSONStore) GetTasks(ctx context.Context, workspaceID int) ([]Task, error) {
s.mu.Lock()
defer s.mu.Unlock()

result := []Task{}

for _, task := range s.tasks {
if task.WorkspaceID == workspaceID {
result = append(result, task)
}
}

return result, nil
}

func (s *JSONStore) CreateTask(ctx context.Context, input Task) (Task, error) {
s.mu.Lock()
defer s.mu.Unlock()

task := Task{
ID:          s.nextID,
WorkspaceID: input.WorkspaceID,
CreatedBy:   input.CreatedBy,
Title:       input.Title,
Status:      input.Status,
Priority:    input.Priority,
Deadline:    input.Deadline,
CreatedAt:   time.Now().Format("2006-01-02 15:04"),
}

s.nextID++
s.tasks = append(s.tasks, task)

return task, s.saveLocked()
}

func (s *JSONStore) UpdateTaskTitle(ctx context.Context, id int, workspaceID int, title string) (Task, bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

for i := range s.tasks {
if s.tasks[i].ID == id && s.tasks[i].WorkspaceID == workspaceID {
s.tasks[i].Title = title
return s.tasks[i], true, s.saveLocked()
}
}

return Task{}, false, nil
}

func (s *JSONStore) UpdateTaskDeadline(ctx context.Context, id int, workspaceID int, deadline string) (Task, bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

for i := range s.tasks {
if s.tasks[i].ID == id && s.tasks[i].WorkspaceID == workspaceID {
s.tasks[i].Deadline = deadline
return s.tasks[i], true, s.saveLocked()
}
}

return Task{}, false, nil
}

func (s *JSONStore) UpdateTaskStatus(ctx context.Context, id int, workspaceID int, status string) (Task, bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

for i := range s.tasks {
if s.tasks[i].ID == id && s.tasks[i].WorkspaceID == workspaceID {
s.tasks[i].Status = status
return s.tasks[i], true, s.saveLocked()
}
}

return Task{}, false, nil
}

func (s *JSONStore) UpdateTaskPriority(ctx context.Context, id int, workspaceID int, priority string) (Task, bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

for i := range s.tasks {
if s.tasks[i].ID == id && s.tasks[i].WorkspaceID == workspaceID {
s.tasks[i].Priority = priority
return s.tasks[i], true, s.saveLocked()
}
}

return Task{}, false, nil
}

func (s *JSONStore) DeleteTask(ctx context.Context, id int, workspaceID int) (bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

for i, task := range s.tasks {
if task.ID == id && task.WorkspaceID == workspaceID {
s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
return true, s.saveLocked()
}
}

return false, nil
}

func (s *JSONStore) load() {
now := time.Now().Format("2006-01-02 15:04")

data, err := os.ReadFile(s.file)
if err != nil {
s.tasks = []Task{
{ID: 1, WorkspaceID: 1, CreatedBy: 1, Title: "Создать первый мини-проект", Status: "done", Priority: "medium", Deadline: "", CreatedAt: now},
{ID: 2, WorkspaceID: 1, CreatedBy: 1, Title: "Добавить PostgreSQL", Status: "done", Priority: "high", Deadline: "", CreatedAt: now},
{ID: 3, WorkspaceID: 1, CreatedBy: 1, Title: "Добавить рабочие области", Status: "todo", Priority: "high", Deadline: "", CreatedAt: now},
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

if s.tasks[i].WorkspaceID == 0 {
s.tasks[i].WorkspaceID = 1
changed = true
}

if s.tasks[i].Status == "" {
s.tasks[i].Status = "todo"
changed = true
}

if s.tasks[i].Priority == "" {
s.tasks[i].Priority = "medium"
changed = true
}

if s.tasks[i].CreatedAt == "" {
s.tasks[i].CreatedAt = now
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
