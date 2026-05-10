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

func (s *PostgresStore) GetTasks(ctx context.Context) ([]Task, error) {
	rows, err := s.db.Query(ctx, `
SELECT id, title, status, priority, COALESCE(deadline::text, ''), to_char(created_at, 'YYYY-MM-DD HH24:MI')
FROM tasks
ORDER BY id;
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []Task{}

	for rows.Next() {
		var task Task

		if err := rows.Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline, &task.CreatedAt); err != nil {
			return nil, err
		}

		result = append(result, task)
	}

	return result, rows.Err()
}

func (s *PostgresStore) CreateTask(ctx context.Context, input Task) (Task, error) {
	var task Task

	err := s.db.QueryRow(ctx, `
INSERT INTO tasks (title, status, priority, deadline)
VALUES ($1, $2, $3, NULLIF($4, '')::date)
RETURNING id, title, status, priority, COALESCE(deadline::text, ''), to_char(created_at, 'YYYY-MM-DD HH24:MI');
`, input.Title, input.Status, input.Priority, input.Deadline).
		Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline, &task.CreatedAt)

	return task, err
}

func (s *PostgresStore) UpdateTaskTitle(ctx context.Context, id int, title string) (Task, bool, error) {
	var task Task

	err := s.db.QueryRow(ctx, `
UPDATE tasks
SET title = $1
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, ''), to_char(created_at, 'YYYY-MM-DD HH24:MI');
`, title, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline, &task.CreatedAt)

	return scanTaskResult(task, err)
}

func (s *PostgresStore) UpdateTaskDeadline(ctx context.Context, id int, deadline string) (Task, bool, error) {
	var task Task

	err := s.db.QueryRow(ctx, `
UPDATE tasks
SET deadline = NULLIF($1, '')::date
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, ''), to_char(created_at, 'YYYY-MM-DD HH24:MI');
`, deadline, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline, &task.CreatedAt)

	return scanTaskResult(task, err)
}

func (s *PostgresStore) UpdateTaskStatus(ctx context.Context, id int, status string) (Task, bool, error) {
	var task Task

	err := s.db.QueryRow(ctx, `
UPDATE tasks
SET status = $1
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, ''), to_char(created_at, 'YYYY-MM-DD HH24:MI');
`, status, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline, &task.CreatedAt)

	return scanTaskResult(task, err)
}

func (s *PostgresStore) UpdateTaskPriority(ctx context.Context, id int, priority string) (Task, bool, error) {
	var task Task

	err := s.db.QueryRow(ctx, `
UPDATE tasks
SET priority = $1
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, ''), to_char(created_at, 'YYYY-MM-DD HH24:MI');
`, priority, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline, &task.CreatedAt)

	return scanTaskResult(task, err)
}

func (s *PostgresStore) DeleteTask(ctx context.Context, id int) (bool, error) {
	result, err := s.db.Exec(ctx, `
DELETE FROM tasks
WHERE id = $1;
`, id)
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
('Добавить дату создания задачи', 'todo', 'high', NULL);
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
