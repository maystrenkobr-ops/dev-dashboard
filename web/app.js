const statuses = ["todo", "in_progress", "done"];
let currentUser = null;
let allTasks = [];
let workspaces = [];
let selectedWorkspaceId = 0;
let draggedTaskId = null;
let membersPanelOpen = false;

function loadCurrentUser() {
fetch("/api/me")
.then(response => {
if (response.status === 401) {
window.location.href = "/login";
return null;
}

return response.json();
})
.then(user => {
if (!user) {
return;
}

currentUser = user;
renderUserPanel();
loadWorkspaces();
});
}

function loadWorkspaces() {
fetch("/api/workspaces")
.then(response => response.json())
.then(data => {
workspaces = data;

const savedWorkspaceID = Number(localStorage.getItem("selected_workspace_id"));
const savedExists = workspaces.some(workspace => workspace.id === savedWorkspaceID);

if (savedWorkspaceID && savedExists) {
selectedWorkspaceId = savedWorkspaceID;
} else if (workspaces.length > 0) {
selectedWorkspaceId = workspaces[0].id;
}

renderWorkspacePanel();
loadTasks();

if (membersPanelOpen) {
loadWorkspaceMembers();
}
});
}

function renderWorkspacePanel() {
const select = document.getElementById("workspaceSelect");

if (!select) {
return;
}

select.innerHTML = workspaces.map(workspace =>
"<option value='" + workspace.id + "'>" + escapeHtml(workspace.name) + "</option>"
).join("");

select.value = String(selectedWorkspaceId);
}

function changeWorkspace() {
const select = document.getElementById("workspaceSelect");
selectedWorkspaceId = Number(select.value);
localStorage.setItem("selected_workspace_id", String(selectedWorkspaceId));
loadTasks();

if (membersPanelOpen) {
loadWorkspaceMembers();
}
}

function createWorkspace() {
const name = prompt("Название рабочей области:");

if (name === null) {
return;
}

if (name.trim() === "") {
alert("Название рабочей области не может быть пустым");
return;
}

fetch("/api/workspaces", {
method: "POST",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
name: name.trim()
})
})
.then(response => response.json().then(data => ({ ok: response.ok, data })))
.then(result => {
if (!result.ok) {
alert(result.data.error || "Не удалось создать рабочую область");
return;
}

selectedWorkspaceId = result.data.id;
localStorage.setItem("selected_workspace_id", String(selectedWorkspaceId));
loadWorkspaces();
});
}

function toggleMembersPanel() {
const panel = document.getElementById("membersPanel");

if (!panel) {
return;
}

membersPanelOpen = panel.classList.contains("hidden");

if (membersPanelOpen) {
panel.classList.remove("hidden");
loadWorkspaceMembers();
} else {
panel.classList.add("hidden");
}
}

function loadWorkspaceMembers() {
if (!selectedWorkspaceId) {
return;
}

fetch("/api/workspaces/" + selectedWorkspaceId + "/members")
.then(response => response.json().then(data => ({ ok: response.ok, data })))
.then(result => {
const list = document.getElementById("membersList");

if (!list) {
return;
}

if (!result.ok) {
list.innerHTML = "<p class='auth-error'>" + escapeHtml(result.data.error || "Не удалось получить участников") + "</p>";
return;
}

renderMembers(result.data);
});
}

function renderMembers(members) {
const list = document.getElementById("membersList");

if (!list) {
return;
}

if (!members || members.length === 0) {
list.innerHTML = "<div class='empty'>Участников пока нет</div>";
return;
}

list.innerHTML = members.map(member =>
"<div class='member-row'>" +
"<div>" +
"<b>" + escapeHtml(member.username || ("user #" + member.user_id)) + "</b>" +
"<br><span>Роль: " + escapeHtml(member.role) + "</span>" +
"<br><span>Добавлен: " + escapeHtml(member.created_at) + "</span>" +
"</div>" +
"<button class='delete-btn' onclick='removeWorkspaceMember(" + member.user_id + ")'>Удалить</button>" +
"</div>"
).join("");
}

function addWorkspaceMember() {
const usernameInput = document.getElementById("memberUsername");
const roleInput = document.getElementById("memberRole");

const username = usernameInput.value.trim();
const role = roleInput.value;

if (username === "") {
alert("Введите логин пользователя");
return;
}

fetch("/api/workspaces/" + selectedWorkspaceId + "/members", {
method: "POST",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
username: username,
role: role
})
})
.then(response => response.json().then(data => ({ ok: response.ok, data })))
.then(result => {
if (!result.ok) {
alert(result.data.error || "Не удалось добавить пользователя");
return;
}

usernameInput.value = "";
loadWorkspaceMembers();
});
}

function removeWorkspaceMember(userID) {
if (!confirm("Удалить пользователя из рабочей области?")) {
return;
}

fetch("/api/workspaces/" + selectedWorkspaceId + "/members/" + userID, {
method: "DELETE"
})
.then(response => response.json().then(data => ({ ok: response.ok, data })))
.then(result => {
if (!result.ok) {
alert(result.data.error || "Не удалось удалить участника");
return;
}

loadWorkspaceMembers();
});
}

function taskQuery() {
return "?workspace_id=" + selectedWorkspaceId;
}

function loadTasks() {
if (!selectedWorkspaceId) {
allTasks = [];
renderTasks();
return;
}

fetch("/tasks" + taskQuery())
.then(response => response.json())
.then(tasks => {
allTasks = Array.isArray(tasks) ? tasks : [];
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
column.innerHTML = "<div class='empty'>Перетащи задачу сюда</div>";
return;
}

column.innerHTML = filtered.map(task => {
const deadlineClass = getDeadlineClass(task);

const createdAtHtml = task.created_at
? "<div class='created-at'>Создано: " + escapeHtml(task.created_at) + "</div>"
: "<div class='created-at'>Создано: неизвестно</div>";

const deadlineHtml = task.deadline
? "<div class='deadline'>Срок: " + escapeHtml(task.deadline) + "</div>"
: "<div class='deadline deadline-empty'>Без срока</div>";

return "<div class='task " + deadlineClass + "' draggable='true' ondragstart='handleDragStart(event, " + task.id + ")' ondragend='handleDragEnd(event)'>" +
"<div class='task-id'>#" + task.id + "</div>" +
"<div class='task-title'>" + escapeHtml(task.title) + "</div>" +
"<div class='priority priority-" + task.priority + "'>" + task.priority + "</div>" +
createdAtHtml +
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

function setupDragAndDrop() {
statuses.forEach(status => {
const taskList = document.getElementById(status);

taskList.addEventListener("dragover", event => {
event.preventDefault();
taskList.classList.add("drag-over");
});

taskList.addEventListener("dragleave", () => {
taskList.classList.remove("drag-over");
});

taskList.addEventListener("drop", event => {
event.preventDefault();
taskList.classList.remove("drag-over");

if (!draggedTaskId) {
return;
}

const task = allTasks.find(item => item.id === draggedTaskId);

if (!task || task.status === status) {
return;
}

updateStatus(draggedTaskId, status);
});
});
}

function handleDragStart(event, id) {
draggedTaskId = id;
event.currentTarget.classList.add("dragging");
event.dataTransfer.effectAllowed = "move";
}

function handleDragEnd(event) {
event.currentTarget.classList.remove("dragging");
draggedTaskId = null;

statuses.forEach(status => {
document.getElementById(status).classList.remove("drag-over");
});
}

function getDeadlineClass(task) {
if (!task.deadline || task.status === "done") {
return "";
}

const today = new Date();
today.setHours(0, 0, 0, 0);

const deadline = new Date(task.deadline + "T00:00:00");

if (deadline < today) {
return "task-overdue";
}

if (deadline.getTime() === today.getTime()) {
return "task-today";
}

return "";
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

fetch("/tasks" + taskQuery(), {
method: "POST",
headers: {
"Content-Type": "application/json"
},
body: JSON.stringify({
workspace_id: selectedWorkspaceId,
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

fetch("/tasks/" + id + "/title" + taskQuery(), {
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

fetch("/tasks/" + id + "/deadline" + taskQuery(), {
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
fetch("/tasks/" + id + "/status" + taskQuery(), {
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
fetch("/tasks/" + id + "/priority" + taskQuery(), {
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
fetch("/tasks/" + id + taskQuery(), {
method: "DELETE"
})
.then(response => response.json())
.then(() => {
loadTasks();
});
}

function renderUserPanel() {
const panel = document.getElementById("user-panel");

if (!panel || !currentUser) {
return;
}

const adminLink = currentUser.role === "admin"
? "<a class='panel-link' href='/admin/users'>Пользователи</a>"
: "";

panel.innerHTML =
"<span class='user-name'>" + escapeHtml(currentUser.username) + " · " + escapeHtml(currentUser.role) + "</span>" +
adminLink +
"<button onclick='logout()'>Выйти</button>";
}

function logout() {
fetch("/api/logout", {
method: "POST"
}).then(() => {
window.location.href = "/login";
});
}

function escapeHtml(text) {
return String(text || "")
.replaceAll("&", "&amp;")
.replaceAll("<", "&lt;")
.replaceAll(">", "&gt;")
.replaceAll('"', "&quot;")
.replaceAll("'", "&#039;");
}

function escapeJs(text) {
return String(text || "")
.replaceAll("\\", "\\\\")
.replaceAll('"', '\\"')
.replaceAll("'", "\\'");
}

setupDragAndDrop();
loadCurrentUser();
