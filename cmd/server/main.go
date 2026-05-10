package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Deadline string `json:"deadline"`
}

var tasks []Task
var nextID = 1
var db *pgxpool.Pool

const tasksFile = "data/tasks.json"

func main() {
	initStorage()

	if db != nil {
		defer db.Close()
	}

	router := gin.Default()

	router.StaticFile("/static/styles.css", "web/styles.css")
	router.StaticFile("/static/app.js", "web/app.js")

	router.GET("/", func(c *gin.Context) {
		c.File("web/index.html")
	})

	router.HEAD("/", func(c *gin.Context) {
		c.Status(200)
	})

	router.GET("/tasks", func(c *gin.Context) {
		result, err := getTasks()
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось получить задачи"})
			return
		}

		c.JSON(200, result)
	})

	router.POST("/tasks", func(c *gin.Context) {
		var input struct {
			Title    string `json:"title"`
			Status   string `json:"status"`
			Priority string `json:"priority"`
			Deadline string `json:"deadline"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Неверный JSON"})
			return
		}

		input.Title = strings.TrimSpace(input.Title)
		input.Deadline = strings.TrimSpace(input.Deadline)

		if input.Title == "" {
			c.JSON(400, gin.H{"error": "Название задачи обязательно"})
			return
		}

		if input.Status == "" {
			input.Status = "todo"
		}

		if input.Priority == "" {
			input.Priority = "medium"
		}

		if !isValidStatus(input.Status) {
			c.JSON(400, gin.H{"error": "Некорректный статус"})
			return
		}

		if !isValidPriority(input.Priority) {
			c.JSON(400, gin.H{"error": "Некорректный приоритет"})
			return
		}

		if !isValidDeadline(input.Deadline) {
			c.JSON(400, gin.H{"error": "Некорректная дата. Используй формат YYYY-MM-DD"})
			return
		}

		task, err := createTask(Task{
			Title:    input.Title,
			Status:   input.Status,
			Priority: input.Priority,
			Deadline: input.Deadline,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось создать задачу"})
			return
		}

		c.JSON(201, task)
	})

	router.PATCH("/tasks/:id/title", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный ID"})
			return
		}

		var input struct {
			Title string `json:"title"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Неверный JSON"})
			return
		}

		input.Title = strings.TrimSpace(input.Title)

		if input.Title == "" {
			c.JSON(400, gin.H{"error": "Название задачи обязательно"})
			return
		}

		task, ok, err := updateTaskTitle(id, input.Title)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось изменить задачу"})
			return
		}

		if !ok {
			c.JSON(404, gin.H{"error": "Задача не найдена"})
			return
		}

		c.JSON(200, task)
	})

	router.PATCH("/tasks/:id/deadline", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный ID"})
			return
		}

		var input struct {
			Deadline string `json:"deadline"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Неверный JSON"})
			return
		}

		input.Deadline = strings.TrimSpace(input.Deadline)

		if !isValidDeadline(input.Deadline) {
			c.JSON(400, gin.H{"error": "Некорректная дата. Используй формат YYYY-MM-DD"})
			return
		}

		task, ok, err := updateTaskDeadline(id, input.Deadline)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось изменить срок"})
			return
		}

		if !ok {
			c.JSON(404, gin.H{"error": "Задача не найдена"})
			return
		}

		c.JSON(200, task)
	})

	router.PATCH("/tasks/:id/status", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный ID"})
			return
		}

		var input struct {
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Неверный JSON"})
			return
		}

		if !isValidStatus(input.Status) {
			c.JSON(400, gin.H{"error": "Некорректный статус"})
			return
		}

		task, ok, err := updateTaskStatus(id, input.Status)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось изменить статус"})
			return
		}

		if !ok {
			c.JSON(404, gin.H{"error": "Задача не найдена"})
			return
		}

		c.JSON(200, task)
	})

	router.PATCH("/tasks/:id/priority", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный ID"})
			return
		}

		var input struct {
			Priority string `json:"priority"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Неверный JSON"})
			return
		}

		if !isValidPriority(input.Priority) {
			c.JSON(400, gin.H{"error": "Некорректный приоритет"})
			return
		}

		task, ok, err := updateTaskPriority(id, input.Priority)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось изменить приоритет"})
			return
		}

		if !ok {
			c.JSON(404, gin.H{"error": "Задача не найдена"})
			return
		}

		c.JSON(200, task)
	})

	router.DELETE("/tasks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный ID"})
			return
		}

		ok, err := deleteTask(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Не удалось удалить задачу"})
			return
		}

		if !ok {
			c.JSON(404, gin.H{"error": "Задача не найдена"})
			return
		}

		c.JSON(200, gin.H{"message": "Задача удалена"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}

func initStorage() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		loadTasks()
		return
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		panic(err)
	}

	db = pool

	if err := createTasksTable(); err != nil {
		panic(err)
	}

	if err := seedTasksIfEmpty(); err != nil {
		panic(err)
	}
}

func createTasksTable() error {
	_, err := db.Exec(context.Background(), `
CREATE TABLE IF NOT EXISTS tasks (
id SERIAL PRIMARY KEY,
title TEXT NOT NULL,
status TEXT NOT NULL DEFAULT 'todo',
priority TEXT NOT NULL DEFAULT 'medium',
deadline DATE NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)
	return err
}

func seedTasksIfEmpty() error {
	var count int
	if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM tasks`).Scan(&count); err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err := db.Exec(context.Background(), `
INSERT INTO tasks (title, status, priority, deadline)
VALUES
('Создать первый мини-проект', 'done', 'medium', NULL),
('Подключить PostgreSQL', 'done', 'high', NULL),
('Довести dashboard до портфолио-уровня', 'todo', 'high', NULL);
`)

	return err
}

func getTasks() ([]Task, error) {
	if db == nil {
		return tasks, nil
	}

	rows, err := db.Query(context.Background(), `
SELECT id, title, status, priority, COALESCE(deadline::text, '')
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

		if err := rows.Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline); err != nil {
			return nil, err
		}

		result = append(result, task)
	}

	return result, rows.Err()
}

func createTask(input Task) (Task, error) {
	if db == nil {
		task := Task{
			ID:       nextID,
			Title:    input.Title,
			Status:   input.Status,
			Priority: input.Priority,
			Deadline: input.Deadline,
		}

		nextID++
		tasks = append(tasks, task)
		saveTasks()

		return task, nil
	}

	var task Task

	err := db.QueryRow(context.Background(), `
INSERT INTO tasks (title, status, priority, deadline)
VALUES ($1, $2, $3, NULLIF($4, '')::date)
RETURNING id, title, status, priority, COALESCE(deadline::text, '');
`, input.Title, input.Status, input.Priority, input.Deadline).
		Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline)

	return task, err
}

func updateTaskTitle(id int, title string) (Task, bool, error) {
	if db == nil {
		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Title = title
				saveTasks()
				return tasks[i], true, nil
			}
		}

		return Task{}, false, nil
	}

	var task Task

	err := db.QueryRow(context.Background(), `
UPDATE tasks
SET title = $1
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, '');
`, title, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Task{}, false, nil
		}

		return Task{}, false, err
	}

	return task, true, nil
}

func updateTaskDeadline(id int, deadline string) (Task, bool, error) {
	if db == nil {
		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Deadline = deadline
				saveTasks()
				return tasks[i], true, nil
			}
		}

		return Task{}, false, nil
	}

	var task Task

	err := db.QueryRow(context.Background(), `
UPDATE tasks
SET deadline = NULLIF($1, '')::date
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, '');
`, deadline, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Task{}, false, nil
		}

		return Task{}, false, err
	}

	return task, true, nil
}

func updateTaskStatus(id int, status string) (Task, bool, error) {
	if db == nil {
		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Status = status
				saveTasks()
				return tasks[i], true, nil
			}
		}

		return Task{}, false, nil
	}

	var task Task

	err := db.QueryRow(context.Background(), `
UPDATE tasks
SET status = $1
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, '');
`, status, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Task{}, false, nil
		}

		return Task{}, false, err
	}

	return task, true, nil
}

func updateTaskPriority(id int, priority string) (Task, bool, error) {
	if db == nil {
		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Priority = priority
				saveTasks()
				return tasks[i], true, nil
			}
		}

		return Task{}, false, nil
	}

	var task Task

	err := db.QueryRow(context.Background(), `
UPDATE tasks
SET priority = $1
WHERE id = $2
RETURNING id, title, status, priority, COALESCE(deadline::text, '');
`, priority, id).Scan(&task.ID, &task.Title, &task.Status, &task.Priority, &task.Deadline)

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Task{}, false, nil
		}

		return Task{}, false, err
	}

	return task, true, nil
}

func deleteTask(id int) (bool, error) {
	if db == nil {
		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks()
				return true, nil
			}
		}

		return false, nil
	}

	result, err := db.Exec(context.Background(), `
DELETE FROM tasks
WHERE id = $1;
`, id)
	if err != nil {
		return false, err
	}

	return result.RowsAffected() > 0, nil
}

func loadTasks() {
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		tasks = []Task{
			{ID: 1, Title: "Создать первый мини-проект", Status: "done", Priority: "medium", Deadline: ""},
			{ID: 2, Title: "Добавить сохранение в JSON", Status: "done", Priority: "high", Deadline: ""},
			{ID: 3, Title: "Добавить PostgreSQL", Status: "todo", Priority: "high", Deadline: ""},
		}
		nextID = 4
		saveTasks()
		return
	}

	if err := json.Unmarshal(data, &tasks); err != nil {
		tasks = []Task{}
		nextID = 1
		return
	}

	changed := false
	nextID = 1

	for i := range tasks {
		if tasks[i].ID >= nextID {
			nextID = tasks[i].ID + 1
		}

		if tasks[i].Priority == "" {
			tasks[i].Priority = "medium"
			changed = true
		}
	}

	if changed {
		saveTasks()
	}
}

func saveTasks() {
	os.MkdirAll("data", 0755)

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return
	}

	os.WriteFile(tasksFile, data, 0644)
}

func isValidStatus(status string) bool {
	return status == "todo" || status == "in_progress" || status == "done"
}

func isValidPriority(priority string) bool {
	return priority == "low" || priority == "medium" || priority == "high"
}

func isValidDeadline(deadline string) bool {
	if deadline == "" {
		return true
	}

	_, err := time.Parse("2006-01-02", deadline)
	return err == nil
}
