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
* {
box-sizing: border-box;
}

body {
font-family: Arial, sans-serif;
background: #0f172a;
color: #f9fafb;
padding: 40px;
margin: 0;
}

.app {
max-width: 1200px;
margin: 0 auto;
}

.header {
background: #1e293b;
padding: 24px;
border-radius: 16px;
margin-bottom: 20px;
}

h1 {
margin: 0 0 10px 0;
font-size: 34px;
}

.subtitle {
color: #cbd5e1;
margin: 0;
}

.form {
background: #1e293b;
padding: 18px;
border-radius: 16px;
margin-bottom: 20px;
display: flex;
gap: 10px;
flex-wrap: wrap;
align-items: center;
}

input, select, button {
padding: 12px;
border-radius: 10px;
border: none;
font-size: 14px;
}

input {
width: 360px;
background: #334155;
color: white;
}

input::placeholder {
color: #94a3b8;
}

select {
background: #334155;
color: white;
}

button {
cursor: pointer;
background: #2563eb;
color: white;
font-weight: bold;
}

button:hover {
background: #1d4ed8;
}

.board {
display: grid;
grid-template-columns: repeat(3, 1fr);
gap: 16px;
}

.column {
background: #1e293b;
border-radius: 16px;
padding: 16px;
min-height: 420px;
}

.column h2 {
margin: 0 0 14px 0;
font-size: 18px;
color: #e2e8f0;
}

.counter {
color: #94a3b8;
font-size: 14px;
font-weight: normal;
}

.task {
background: #334155;
padding: 14px;
border-radius: 12px;
margin-bottom: 12px;
}

.task-title {
font-size: 16px;
font-weight: bold;
margin-bottom: 8px;
}

.task-id {
color: #93c5fd;
font-size: 13px;
margin-bottom: 10px;
}

.actions {
display: flex;
gap: 6px;
flex-wrap: wrap;
margin-top: 10px;
}

.small-btn {
font-size: 12px;
padding: 8px;
background: #475569;
}

.small-btn:hover {
background: #64748b;
}

.delete-btn {
background: #dc2626;
}

.delete-btn:hover {
background: #b91c1c;
}

.empty {
color: #94a3b8;
font-size: 14px;
padding: 10px;
border: 1px dashed #475569;
border-radius: 10px;
}

@media (max-width: 900px) {
.board {
grid-template-columns: 1fr;
}

input {
width: 100%;
}
}
</style>
</head>
<body>
<div class="app">
<div class="header">
<h1>Dev Dashboard</h1>
<p class="subtitle">Мини-доска задач на Go + Gin + JSON.</p>
</div>

<div class="form">
<input id="title" placeholder="Новая задача" />

<select id="status">
<option value="todo">todo</option>
<option value="in_progress">in_progress</option>
<option value="done">done</option>
</select>

<button onclick="createTask()">Добавить</button>
</div>

<div class="board">
<div class="column">
<h2>TODO <span class="counter" id="todo-count"></span></h2>
<div id="todo"></div>
</div>

<div class="column">
<h2>IN PROGRESS <span class="counter" id="in_progress-count"></span></h2>
<div id="in_progress"></div>
</div>

<div class="column">
<h2>DONE <span class="counter" id="done-count"></span></h2>
<div id="done"></div>
</div>
</div>
</div>

<script>
const statuses = ["todo", "in_progress", "done"];

function loadTasks() {
fetch("/tasks")
.then(response => response.json())
.then(tasks => {
statuses.forEach(status => {
const column = document.getElementById(status);
const count = document.getElementById(status + "-count");
const filtered = tasks.filter(task => task.status === status);

count.innerText = "(" + filtered.length + ")";

if (filtered.length === 0) {
column.innerHTML = "<div class='empty'>Пока пусто</div>";
return;
}

column.innerHTML = filtered.map(task =>
"<div class='task'>" +
"<div class='task-id'>#" + task.id + "</div>" +
"<div class='task-title'>" + escapeHtml(task.title) + "</div>" +
"<div class='actions'>" +
"<button class='small-btn' onclick='updateStatus(" + task.id + ", \"todo\")'>todo</button>" +
"<button class='small-btn' onclick='updateStatus(" + task.id + ", \"in_progress\")'>progress</button>" +
"<button class='small-btn' onclick='updateStatus(" + task.id + ", \"done\")'>done</button>" +
"<button class='small-btn delete-btn' onclick='deleteTask(" + task.id + ")'>Удалить</button>" +
"</div>" +
"</div>"
).join("");
});
});
}

function createTask() {
const titleInput = document.getElementById("title");
const title = titleInput.value;
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
titleInput.value = "";
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

function escapeHtml(text) {
return text
.replaceAll("&", "&amp;")
.replaceAll("<", "&lt;")
.replaceAll(">", "&gt;")
.replaceAll('"', "&quot;")
.replaceAll("'", "&#039;");
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
			{ID: 3, Title: "Добавить доску задач", Status: "todo"},
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
