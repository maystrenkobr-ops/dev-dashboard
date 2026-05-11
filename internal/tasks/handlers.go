package tasks

import (
"net/http"
"strconv"
"strings"

"github.com/gin-gonic/gin"
"github.com/maystrenkobr-ops/dev-dashboard/internal/auth"
"github.com/maystrenkobr-ops/dev-dashboard/internal/workspaces"
)

type Handler struct {
store          Store
workspaceStore workspaces.Store
}

func RegisterRoutes(router gin.IRoutes, store Store, workspaceStore workspaces.Store) {
handler := &Handler{
store:          store,
workspaceStore: workspaceStore,
}

router.GET("", handler.getTasks)
router.POST("", handler.createTask)
router.PATCH("/:id/title", handler.updateTaskTitle)
router.PATCH("/:id/deadline", handler.updateTaskDeadline)
router.PATCH("/:id/status", handler.updateTaskStatus)
router.PATCH("/:id/priority", handler.updateTaskPriority)
router.DELETE("/:id", handler.deleteTask)
}

func (h *Handler) getTasks(c *gin.Context) {
user, workspaceID, ok := h.resolveWorkspace(c, 0)
if !ok {
return
}

result, err := h.store.GetTasks(c.Request.Context(), workspaceID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить задачи"})
return
}

_ = user
c.JSON(http.StatusOK, result)
}

func (h *Handler) createTask(c *gin.Context) {
var input struct {
WorkspaceID int    `json:"workspace_id"`
Title       string `json:"title"`
Status      string `json:"status"`
Priority    string `json:"priority"`
Deadline    string `json:"deadline"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

user, workspaceID, ok := h.resolveWorkspace(c, input.WorkspaceID)
if !ok {
return
}

input.Title = strings.TrimSpace(input.Title)
input.Deadline = strings.TrimSpace(input.Deadline)

if input.Title == "" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Название задачи обязательно"})
return
}

if input.Status == "" {
input.Status = "todo"
}

if input.Priority == "" {
input.Priority = "medium"
}

if !IsValidStatus(input.Status) {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный статус"})
return
}

if !IsValidPriority(input.Priority) {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный приоритет"})
return
}

if !IsValidDeadline(input.Deadline) {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректная дата. Используй формат YYYY-MM-DD"})
return
}

task, err := h.store.CreateTask(c.Request.Context(), Task{
WorkspaceID: workspaceID,
CreatedBy:   user.ID,
Title:       input.Title,
Status:      input.Status,
Priority:    input.Priority,
Deadline:    input.Deadline,
})
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать задачу"})
return
}

c.JSON(http.StatusCreated, task)
}

func (h *Handler) updateTaskTitle(c *gin.Context) {
_, workspaceID, ok := h.resolveWorkspace(c, 0)
if !ok {
return
}

id, ok := parseID(c)
if !ok {
return
}

var input struct {
Title string `json:"title"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

input.Title = strings.TrimSpace(input.Title)

if input.Title == "" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Название задачи обязательно"})
return
}

task, found, err := h.store.UpdateTaskTitle(c.Request.Context(), id, workspaceID, input.Title)
writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить задачу")
}

func (h *Handler) updateTaskDeadline(c *gin.Context) {
_, workspaceID, ok := h.resolveWorkspace(c, 0)
if !ok {
return
}

id, ok := parseID(c)
if !ok {
return
}

var input struct {
Deadline string `json:"deadline"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

input.Deadline = strings.TrimSpace(input.Deadline)

if !IsValidDeadline(input.Deadline) {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректная дата. Используй формат YYYY-MM-DD"})
return
}

task, found, err := h.store.UpdateTaskDeadline(c.Request.Context(), id, workspaceID, input.Deadline)
writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить срок")
}

func (h *Handler) updateTaskStatus(c *gin.Context) {
_, workspaceID, ok := h.resolveWorkspace(c, 0)
if !ok {
return
}

id, ok := parseID(c)
if !ok {
return
}

var input struct {
Status string `json:"status"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

if !IsValidStatus(input.Status) {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный статус"})
return
}

task, found, err := h.store.UpdateTaskStatus(c.Request.Context(), id, workspaceID, input.Status)
writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить статус")
}

func (h *Handler) updateTaskPriority(c *gin.Context) {
_, workspaceID, ok := h.resolveWorkspace(c, 0)
if !ok {
return
}

id, ok := parseID(c)
if !ok {
return
}

var input struct {
Priority string `json:"priority"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

if !IsValidPriority(input.Priority) {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный приоритет"})
return
}

task, found, err := h.store.UpdateTaskPriority(c.Request.Context(), id, workspaceID, input.Priority)
writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить приоритет")
}

func (h *Handler) deleteTask(c *gin.Context) {
_, workspaceID, ok := h.resolveWorkspace(c, 0)
if !ok {
return
}

id, ok := parseID(c)
if !ok {
return
}

found, err := h.store.DeleteTask(c.Request.Context(), id, workspaceID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить задачу"})
return
}

if !found {
c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Задача удалена"})
}

func (h *Handler) resolveWorkspace(c *gin.Context, requestedWorkspaceID int) (auth.User, int, bool) {
user, ok := auth.CurrentUser(c)
if !ok {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
return auth.User{}, 0, false
}

workspaceID := requestedWorkspaceID

if workspaceID == 0 {
fromQuery, err := strconv.Atoi(c.Query("workspace_id"))
if err == nil {
workspaceID = fromQuery
}
}

if workspaceID == 0 {
personalWorkspace, err := h.workspaceStore.EnsurePersonalWorkspace(c.Request.Context(), user.ID, user.Username)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить личную область"})
return auth.User{}, 0, false
}

workspaceID = personalWorkspace.ID
}

canAccess, err := h.workspaceStore.UserCanAccessWorkspace(c.Request.Context(), workspaceID, user.ID, user.Role == "admin")
if err != nil || !canAccess {
c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к рабочей области"})
return auth.User{}, 0, false
}

return user, workspaceID, true
}

func parseID(c *gin.Context) (int, bool) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
return 0, false
}

return id, true
}

func writeTaskUpdateResponse(c *gin.Context, task Task, found bool, err error, errorMessage string) {
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage})
return
}

if !found {
c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
return
}

c.JSON(http.StatusOK, task)
}
