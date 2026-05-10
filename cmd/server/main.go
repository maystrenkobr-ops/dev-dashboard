package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

const tasksFile = "data/tasks.json"

func main() {
	loadTasks()

	router := gin.Default()

	router.StaticFile("/static/styles.css", "web/styles.css")
	router.StaticFile("/static/app.js", "web/app.js")

	router.GET("/", func(c *gin.Context) {
		c.File("web/index.html")
	})

	router.GET("/tasks", func(c *gin.Context) {
		c.JSON(200, tasks)
	})

	router.POST("/tasks", func(c *gin.Context) {
		var input struct {
			Title    string `json:"title"`
			Status   string `json:"status"`
			Priority string `json:"priority"`
			Deadline string `json:"deadline"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "РќРµРІРµСЂРЅС‹Р№ JSON"})
			return
		}

		input.Title = strings.TrimSpace(input.Title)
		input.Deadline = strings.TrimSpace(input.Deadline)

		if input.Title == "" {
			c.JSON(400, gin.H{"error": "РќР°Р·РІР°РЅРёРµ Р·Р°РґР°С‡Рё РѕР±СЏР·Р°С‚РµР»СЊРЅРѕ"})
			return
		}

		if input.Status == "" {
			input.Status = "todo"
		}

		if input.Priority == "" {
			input.Priority = "medium"
		}

		if !isValidStatus(input.Status) {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ СЃС‚Р°С‚СѓСЃ"})
			return
		}

		if !isValidPriority(input.Priority) {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РїСЂРёРѕСЂРёС‚РµС‚"})
			return
		}

		if !isValidDeadline(input.Deadline) {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅР°СЏ РґР°С‚Р°. РСЃРїРѕР»СЊР·СѓР№ С„РѕСЂРјР°С‚ YYYY-MM-DD"})
			return
		}

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

		c.JSON(201, task)
	})

	router.PATCH("/tasks/:id/title", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ ID"})
			return
		}

		var input struct {
			Title string `json:"title"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "РќРµРІРµСЂРЅС‹Р№ JSON"})
			return
		}

		input.Title = strings.TrimSpace(input.Title)

		if input.Title == "" {
			c.JSON(400, gin.H{"error": "РќР°Р·РІР°РЅРёРµ Р·Р°РґР°С‡Рё РѕР±СЏР·Р°С‚РµР»СЊРЅРѕ"})
			return
		}

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Title = input.Title
				saveTasks()
				c.JSON(200, tasks[i])
				return
			}
		}

		c.JSON(404, gin.H{"error": "Р—Р°РґР°С‡Р° РЅРµ РЅР°Р№РґРµРЅР°"})
	})

	router.PATCH("/tasks/:id/deadline", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ ID"})
			return
		}

		var input struct {
			Deadline string `json:"deadline"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "РќРµРІРµСЂРЅС‹Р№ JSON"})
			return
		}

		input.Deadline = strings.TrimSpace(input.Deadline)

		if !isValidDeadline(input.Deadline) {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅР°СЏ РґР°С‚Р°. РСЃРїРѕР»СЊР·СѓР№ С„РѕСЂРјР°С‚ YYYY-MM-DD"})
			return
		}

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Deadline = input.Deadline
				saveTasks()
				c.JSON(200, tasks[i])
				return
			}
		}

		c.JSON(404, gin.H{"error": "Р—Р°РґР°С‡Р° РЅРµ РЅР°Р№РґРµРЅР°"})
	})

	router.PATCH("/tasks/:id/status", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ ID"})
			return
		}

		var input struct {
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "РќРµРІРµСЂРЅС‹Р№ JSON"})
			return
		}

		if !isValidStatus(input.Status) {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ СЃС‚Р°С‚СѓСЃ"})
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

		c.JSON(404, gin.H{"error": "Р—Р°РґР°С‡Р° РЅРµ РЅР°Р№РґРµРЅР°"})
	})

	router.PATCH("/tasks/:id/priority", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ ID"})
			return
		}

		var input struct {
			Priority string `json:"priority"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "РќРµРІРµСЂРЅС‹Р№ JSON"})
			return
		}

		if !isValidPriority(input.Priority) {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РїСЂРёРѕСЂРёС‚РµС‚"})
			return
		}

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Priority = input.Priority
				saveTasks()
				c.JSON(200, tasks[i])
				return
			}
		}

		c.JSON(404, gin.H{"error": "Р—Р°РґР°С‡Р° РЅРµ РЅР°Р№РґРµРЅР°"})
	})

	router.DELETE("/tasks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ ID"})
			return
		}

		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks()
				c.JSON(200, gin.H{"message": "Р—Р°РґР°С‡Р° СѓРґР°Р»РµРЅР°"})
				return
			}
		}

		c.JSON(404, gin.H{"error": "Р—Р°РґР°С‡Р° РЅРµ РЅР°Р№РґРµРЅР°"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router.Run(":" + port)
}

func loadTasks() {
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		tasks = []Task{
			{ID: 1, Title: "РЎРѕР·РґР°С‚СЊ РїРµСЂРІС‹Р№ РјРёРЅРё-РїСЂРѕРµРєС‚", Status: "done", Priority: "medium", Deadline: ""},
			{ID: 2, Title: "Р”РѕР±Р°РІРёС‚СЊ СЃРѕС…СЂР°РЅРµРЅРёРµ РІ JSON", Status: "done", Priority: "high", Deadline: ""},
			{ID: 3, Title: "Р”РѕР±Р°РІРёС‚СЊ СЃСЂРѕРє Р·Р°РґР°С‡Рё", Status: "todo", Priority: "high", Deadline: ""},
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
