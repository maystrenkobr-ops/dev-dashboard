package tasks

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store Store
}

func RegisterRoutes(router gin.IRoutes, store Store) {
	handler := &Handler{store: store}

	router.GET("", handler.getTasks)
	router.POST("", handler.createTask)
	router.PATCH("/:id/title", handler.updateTaskTitle)
	router.PATCH("/:id/deadline", handler.updateTaskDeadline)
	router.PATCH("/:id/status", handler.updateTaskStatus)
	router.PATCH("/:id/priority", handler.updateTaskPriority)
	router.DELETE("/:id", handler.deleteTask)
}

func (h *Handler) getTasks(c *gin.Context) {
	result, err := h.store.GetTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить задачи"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) createTask(c *gin.Context) {
	var input struct {
		Title    string `json:"title"`
		Status   string `json:"status"`
		Priority string `json:"priority"`
		Deadline string `json:"deadline"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
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
		Title:    input.Title,
		Status:   input.Status,
		Priority: input.Priority,
		Deadline: input.Deadline,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать задачу"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *Handler) updateTaskTitle(c *gin.Context) {
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

	task, found, err := h.store.UpdateTaskTitle(c.Request.Context(), id, input.Title)
	writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить задачу")
}

func (h *Handler) updateTaskDeadline(c *gin.Context) {
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

	task, found, err := h.store.UpdateTaskDeadline(c.Request.Context(), id, input.Deadline)
	writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить срок")
}

func (h *Handler) updateTaskStatus(c *gin.Context) {
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

	task, found, err := h.store.UpdateTaskStatus(c.Request.Context(), id, input.Status)
	writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить статус")
}

func (h *Handler) updateTaskPriority(c *gin.Context) {
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

	task, found, err := h.store.UpdateTaskPriority(c.Request.Context(), id, input.Priority)
	writeTaskUpdateResponse(c, task, found, err, "Не удалось изменить приоритет")
}

func (h *Handler) deleteTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	found, err := h.store.DeleteTask(c.Request.Context(), id)
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
