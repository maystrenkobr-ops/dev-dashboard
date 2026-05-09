package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

var tasks []Task
var nextID = 1

const tasksFile = "data/tasks.json"

func main() {
	loadTasks()

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.File("web/index.html")
	})

	router.GET("/tasks", func(c *gin.Context) {
		c.JSON(200, tasks)
	})

	router.POST("/tasks", func(c *gin.Context) {
		var input struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Неверный JSON"})
			return
		}

		if input.Title == "" {
			c.JSON(400, gin.H{"error": "Название задачи обязательно"})
			return
		}

		if input.Status == "" {
			input.Status = "todo"
		}

		if !isValidStatus(input.Status) {
			c.JSON(400, gin.H{"error": "Некорректный статус"})
			return
		}

		task := Task{
			ID:     nextID,
			Title:  input.Title,
			Status: input.Status,
		}

		nextID++
		tasks = append(tasks, task)

		saveTasks()

		c.JSON(201, task)
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

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Status = input.Status
				saveTasks()
				c.JSON(200, tasks[i])
				return
			}
		}

		c.JSON(404, gin.H{"error": "Задача не найдена"})
	})

	router.DELETE("/tasks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Некорректный ID"})
			return
		}

		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks()
				c.JSON(200, gin.H{"message": "Задача удалена"})
				return
			}
		}

		c.JSON(404, gin.H{"error": "Задача не найдена"})
	})

	router.Run(":8080")
}

func loadTasks() {
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		tasks = []Task{
			{ID: 1, Title: "Создать первый мини-проект", Status: "done"},
			{ID: 2, Title: "Добавить сохранение в JSON", Status: "done"},
			{ID: 3, Title: "Вынести HTML в отдельный файл", Status: "todo"},
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

	nextID = 1

	for _, task := range tasks {
		if task.ID >= nextID {
			nextID = task.ID + 1
		}
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
