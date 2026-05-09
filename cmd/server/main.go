package main

import (
	"encoding/json"
	"net/http"
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
		html := `
<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<title>Dev Dashboard</title>
<style>
body {
font-family: Arial, sans-serif;
background: #111827;
color: #f9fafb;
padding: 40px;
}

.card {
background: #1f2937;
padding: 20px;
border-radius: 12px;
max-width: 820px;
}

h1 {
margin-top: 0;
}

.task {
padding: 12px;
margin: 10px 0;
background: #374151;
border-radius: 8px;
}

.status {
color: #93c5fd;
font-weight: bold;
}

input, select, button {
padding: 10px;
margin: 6px 4px 12px 0;
border-radius: 8px;
border: none;
}

input {
width: 300px;
}

button {
cursor: pointer;
background: #2563eb;
color: white;
}

button:hover {
background: #1d4ed8;
}

.delete-btn {
background: #dc2626;
}

.delete-btn:hover {
background: #b91c1c;
}

.status-btn {
background: #4b5563;
}

.status-btn:hover {
background: #6b7280;
}
</style>
</head>
<body>
<div class="card">
<h1>Dev Dashboard</h1>
<p>Задачи сохраняются в файл <b>data/tasks.json</b>.</p>

<input id="title" placeholder="Новая задача" />

<select id="status">
<option value="todo">todo</option>
<option value="in_progress">in_progress</option>
<option value="done">done</option>
</select>

<button onclick="createTask()">Добавить</button>

<hr>

<div id="tasks"></div>
</div>

<script>
function loadTasks() {
fetch("/tasks")
.then(response => response.json())
.then(tasks => {
const container = document.getElementById("tasks");

if (tasks.length === 0) {
container.innerHTML = "<p>Задач пока нет.</p>";
return;
}

container.innerHTML = tasks.map(task =>
"<div class='task'>" +
"<b>#" + task.id + "</b> " + task.title +
"<br><span class='status'>Статус: " + task.status + "</span>" +
"<br><br>" +
"<button class='status-btn' onclick='updateStatus(" + task.id + ", \"todo\")'>todo</button>" +
"<button class='status-btn' onclick='updateStatus(" + task.id + ", \"in_progress\")'>in_progress</button>" +
"<button class='status-btn' onclick='updateStatus(" + task.id + ", \"done\")'>done</button>" +
"<button class='delete-btn' onclick='deleteTask(" + task.id + ")'>Удалить</button>" +
"</div>"
).join("");
});
}

function createTask() {
const title = document.getElementById("title").value;
const status = document.getElementById("status").value;

if (title.trim() === "") {
alert("Введите название задачи");
return;
}

fetch("/tasks", {
method: "POST",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
title: title,
status: status
})
})
.then(response => response.json())
.then(() => {
document.getElementById("title").value = "";
loadTasks();
});
}

function updateStatus(id, status) {
fetch("/tasks/" + id + "/status", {
method: "PATCH",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
status: status
})
})
.then(response => response.json())
.then(() => {
loadTasks();
});
}

function deleteTask(id) {
fetch("/tasks/" + id, {
method: "DELETE"
})
.then(response => response.json())
.then(() => {
loadTasks();
});
}

loadTasks();
</script>
</body>
</html>
`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	router.GET("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, tasks)
	})

	router.POST("/tasks", func(c *gin.Context) {
		var input struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
			return
		}

		if input.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Название задачи обязательно"})
			return
		}

		if input.Status == "" {
			input.Status = "todo"
		}

		if !isValidStatus(input.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный статус"})
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

		c.JSON(http.StatusCreated, task)
	})

	router.PATCH("/tasks/:id/status", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
			return
		}

		var input struct {
			Status string `json:"status"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
			return
		}

		if !isValidStatus(input.Status) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный статус"})
			return
		}

		for i := range tasks {
			if tasks[i].ID == id {
				tasks[i].Status = input.Status
				saveTasks()
				c.JSON(http.StatusOK, tasks[i])
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
	})

	router.DELETE("/tasks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
			return
		}

		for i, task := range tasks {
			if task.ID == id {
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks()
				c.JSON(http.StatusOK, gin.H{"message": "Задача удалена"})
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
	})

	router.Run(":8080")
}

func loadTasks() {
	data, err := os.ReadFile(tasksFile)
	if err != nil {
		tasks = []Task{
			{ID: 1, Title: "Создать первый мини-проект", Status: "done"},
			{ID: 2, Title: "Добавить сохранение в JSON", Status: "done"},
			{ID: 3, Title: "Добавить смену статуса", Status: "todo"},
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
