const statuses = ["todo", "in_progress", "done"];
let allTasks = [];

function loadTasks() {
fetch("/tasks")
.then(response => response.json())
.then(tasks => {
allTasks = tasks;
renderTasks();
});
}

function renderTasks() {
const searchInput = document.getElementById("search");
const priorityFilter = document.getElementById("priorityFilter");

const searchText = searchInput ? searchInput.value.toLowerCase().trim() : "";
const selectedPriority = priorityFilter ? priorityFilter.value : "all";

const visibleTasks = allTasks.filter(task => {
const matchesSearch = task.title.toLowerCase().includes(searchText);
const matchesPriority = selectedPriority === "all" || task.priority === selectedPriority;

return matchesSearch && matchesPriority;
});

statuses.forEach(status => {
const column = document.getElementById(status);
const count = document.getElementById(status + "-count");
const filtered = visibleTasks.filter(task => task.status === status);

count.innerText = "(" + filtered.length + ")";

if (filtered.length === 0) {
column.innerHTML = "<div class='empty'>Пока пусто</div>";
return;
}

column.innerHTML = filtered.map(task => {
const deadlineHtml = task.deadline
? "<div class='deadline'>Срок: " + escapeHtml(task.deadline) + "</div>"
: "<div class='deadline deadline-empty'>Без срока</div>";

return "<div class='task'>" +
"<div class='task-id'>#" + task.id + "</div>" +
"<div class='task-title'>" + escapeHtml(task.title) + "</div>" +
"<div class='priority priority-" + task.priority + "'>" + task.priority + "</div>" +
deadlineHtml +
"<div class='actions'>" +
"<button class='small-btn' onclick='updateStatus(" + task.id + ", \"todo\")'>todo</button>" +
"<button class='small-btn' onclick='updateStatus(" + task.id + ", \"in_progress\")'>progress</button>" +
"<button class='small-btn' onclick='updateStatus(" + task.id + ", \"done\")'>done</button>" +
"<button class='small-btn' onclick='updatePriority(" + task.id + ", \"low\")'>low</button>" +
"<button class='small-btn' onclick='updatePriority(" + task.id + ", \"medium\")'>medium</button>" +
"<button class='small-btn' onclick='updatePriority(" + task.id + ", \"high\")'>high</button>" +
"<button class='small-btn' onclick='editTask(" + task.id + ", \"" + escapeJs(task.title) + "\")'>Изменить</button>" +
"<button class='small-btn' onclick='editDeadline(" + task.id + ", \"" + escapeJs(task.deadline || "") + "\")'>Срок</button>" +
"<button class='small-btn delete-btn' onclick='deleteTask(" + task.id + ")'>Удалить</button>" +
"</div>" +
"</div>";
}).join("");
});
}

function createTask() {
const titleInput = document.getElementById("title");
const deadlineInput = document.getElementById("deadline");

const title = titleInput.value;
const status = document.getElementById("status").value;
const priority = document.getElementById("priority").value;
const deadline = deadlineInput.value;

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
status: status,
priority: priority,
deadline: deadline
})
})
.then(response => response.json())
.then(() => {
titleInput.value = "";
deadlineInput.value = "";
loadTasks();
});
}

function editTask(id, currentTitle) {
const newTitle = prompt("Новое название задачи:", currentTitle);

if (newTitle === null) {
return;
}

if (newTitle.trim() === "") {
alert("Название задачи не может быть пустым");
return;
}

fetch("/tasks/" + id + "/title", {
method: "PATCH",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
title: newTitle
})
})
.then(response => response.json())
.then(() => {
loadTasks();
});
}

function editDeadline(id, currentDeadline) {
const newDeadline = prompt("Новый срок в формате YYYY-MM-DD. Чтобы очистить срок — оставь пустым:", currentDeadline);

if (newDeadline === null) {
return;
}

fetch("/tasks/" + id + "/deadline", {
method: "PATCH",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
deadline: newDeadline.trim()
})
})
.then(response => response.json())
.then(data => {
if (data.error) {
alert(data.error);
return;
}

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

function updatePriority(id, priority) {
fetch("/tasks/" + id + "/priority", {
method: "PATCH",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
priority: priority
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

function escapeJs(text) {
return text
.replaceAll("\\", "\\\\")
.replaceAll('"', '\\"')
.replaceAll("'", "\\'");
}

loadTasks();
