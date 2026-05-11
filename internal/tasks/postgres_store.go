package tasks

import (
"context"
"errors"

"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
db *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
pool, err := pgxpool.New(ctx, databaseURL)
if err != nil {
return nil, err
}

store := &PostgresStore{db: pool}

if err := store.createTasksTable(ctx); err != nil {
pool.Close()
return nil, err
}

if err := store.seedTasksIfEmpty(ctx); err != nil {
pool.Close()
return nil, err
}

return store, nil
}

func (s *PostgresStore) Close() {
s.db.Close()
}

func (s *PostgresStore) EnsureWorkspaceSupport(ctx context.Context, defaultWorkspaceID int, defaultUserID int) error {
if defaultWorkspaceID > 0 {
if _, err := s.db.Exec(ctx, `
UPDATE tasks
SET workspace_id = $1
WHERE workspace_id IS NULL;
`, defaultWorkspaceID); err != nil {
return err
}
}

if defaultUserID > 0 {
if _, err := s.db.Exec(ctx, `
UPDATE tasks
SET created_by = $1
WHERE created_by IS NULL;
`, defaultUserID); err != nil {
return err
}
}

return nil
}

func (s *PostgresStore) GetTasks(ctx context.Context, workspaceID int) ([]Task, error) {
rows, err := s.db.Query(ctx, `
SELECT
id,
COALESCE(workspace_id, 0),
COALESCE(created_by, 0),
title,
status,
priority,
COALESCE(deadline::text, ''),
to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM tasks
WHERE workspace_id = $1
ORDER BY id;
`, workspaceID)
if err != nil {
return nil, err
}
defer rows.Close()

result := []Task{}

for rows.Next() {
var task Task

if err := rows.Scan(
&task.ID,
&task.WorkspaceID,
&task.CreatedBy,
&task.Title,
&task.Status,
&task.Priority,
&task.Deadline,
&task.CreatedAt,
); err != nil {
return nil, err
}

result = append(result, task)
}

return result, rows.Err()
}

func (s *PostgresStore) CreateTask(ctx context.Context, input Task) (Task, error) {
var task Task

err := s.db.QueryRow(ctx, `
INSERT INTO tasks (workspace_id, created_by, title, status, priority, deadline)
VALUES ($1, NULLIF($2, 0), $3, $4, $5, NULLIF($6, '')::date)
RETURNING
id,
COALESCE(workspace_id, 0),
COALESCE(created_by, 0),
title,
status,
priority,
COALESCE(deadline::text, ''),
to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI');
`, input.WorkspaceID, input.CreatedBy, input.Title, input.Status, input.Priority, input.Deadline).
Scan(
&task.ID,
&task.WorkspaceID,
&task.CreatedBy,
&task.Title,
&task.Status,
&task.Priority,
&task.Deadline,
&task.CreatedAt,
)

return task, err
}

func (s *PostgresStore) UpdateTaskTitle(ctx context.Context, id int, workspaceID int, title string) (Task, bool, error) {
var task Task

err := s.db.QueryRow(ctx, `
UPDATE tasks
SET title = $1
WHERE id = $2 AND workspace_id = $3
RETURNING
id,
COALESCE(workspace_id, 0),
COALESCE(created_by, 0),
title,
status,
priority,
COALESCE(deadline::text, ''),
to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI');
`, title, id, workspaceID).Scan(
&task.ID,
&task.WorkspaceID,
&task.CreatedBy,
&task.Title,
&task.Status,
&task.Priority,
&task.Deadline,
&task.CreatedAt,
)

return scanTaskResult(task, err)
}

func (s *PostgresStore) UpdateTaskDeadline(ctx context.Context, id int, workspaceID int, deadline string) (Task, bool, error) {
var task Task

err := s.db.QueryRow(ctx, `
UPDATE tasks
SET deadline = NULLIF($1, '')::date
WHERE id = $2 AND workspace_id = $3
RETURNING
id,
COALESCE(workspace_id, 0),
COALESCE(created_by, 0),
title,
status,
priority,
COALESCE(deadline::text, ''),
to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI');
`, deadline, id, workspaceID).Scan(
&task.ID,
&task.WorkspaceID,
&task.CreatedBy,
&task.Title,
&task.Status,
&task.Priority,
&task.Deadline,
&task.CreatedAt,
)

return scanTaskResult(task, err)
}

func (s *PostgresStore) UpdateTaskStatus(ctx context.Context, id int, workspaceID int, status string) (Task, bool, error) {
var task Task

err := s.db.QueryRow(ctx, `
UPDATE tasks
SET status = $1
WHERE id = $2 AND workspace_id = $3
RETURNING
id,
COALESCE(workspace_id, 0),
COALESCE(created_by, 0),
title,
status,
priority,
COALESCE(deadline::text, ''),
to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI');
`, status, id, workspaceID).Scan(
&task.ID,
&task.WorkspaceID,
&task.CreatedBy,
&task.Title,
&task.Status,
&task.Priority,
&task.Deadline,
&task.CreatedAt,
)

return scanTaskResult(task, err)
}

func (s *PostgresStore) UpdateTaskPriority(ctx context.Context, id int, workspaceID int, priority string) (Task, bool, error) {
var task Task

err := s.db.QueryRow(ctx, `
UPDATE tasks
SET priority = $1
WHERE id = $2 AND workspace_id = $3
RETURNING
id,
COALESCE(workspace_id, 0),
COALESCE(created_by, 0),
title,
status,
priority,
COALESCE(deadline::text, ''),
to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI');
`, priority, id, workspaceID).Scan(
&task.ID,
&task.WorkspaceID,
&task.CreatedBy,
&task.Title,
&task.Status,
&task.Priority,
&task.Deadline,
&task.CreatedAt,
)

return scanTaskResult(task, err)
}

func (s *PostgresStore) DeleteTask(ctx context.Context, id int, workspaceID int) (bool, error) {
result, err := s.db.Exec(ctx, `
DELETE FROM tasks
WHERE id = $1 AND workspace_id = $2;
`, id, workspaceID)
if err != nil {
return false, err
}

return result.RowsAffected() > 0, nil
}

func (s *PostgresStore) createTasksTable(ctx context.Context) error {
_, err := s.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS tasks (
id SERIAL PRIMARY KEY,
title TEXT NOT NULL,
status TEXT NOT NULL DEFAULT 'todo',
priority TEXT NOT NULL DEFAULT 'medium',
deadline DATE NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS workspace_id INTEGER NULL REFERENCES workspaces(id) ON DELETE CASCADE;

ALTER TABLE tasks
ADD COLUMN IF NOT EXISTS created_by INTEGER NULL REFERENCES app_users(id) ON DELETE SET NULL;
`)

return err
}

func (s *PostgresStore) seedTasksIfEmpty(ctx context.Context) error {
var count int

if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM tasks`).Scan(&count); err != nil {
return err
}

if count > 0 {
return nil
}

_, err := s.db.Exec(ctx, `
INSERT INTO tasks (title, status, priority, deadline)
VALUES
('Создать первый мини-проект', 'done', 'medium', NULL),
('Подключить PostgreSQL', 'done', 'high', NULL),
('Добавить рабочие области', 'todo', 'high', NULL);
`)

return err
}

func scanTaskResult(task Task, err error) (Task, bool, error) {
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return Task{}, false, nil
}

return Task{}, false, err
}

return task, true, nil
}
