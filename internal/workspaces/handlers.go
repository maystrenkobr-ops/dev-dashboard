package workspaces

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/maystrenkobr-ops/dev-dashboard/internal/auth"
)

type Handler struct {
store     Store
authStore auth.Store
}

func RegisterRoutes(router gin.IRoutes, store Store, authStore auth.Store) {
handler := &Handler{
store:     store,
authStore: authStore,
}

router.GET("/workspaces", handler.listWorkspaces)
router.POST("/workspaces", handler.createWorkspace)
router.GET("/workspaces/:id/members", handler.listMembers)
router.POST("/workspaces/:id/members", handler.addMember)
router.DELETE("/workspaces/:id/members/:userID", handler.removeMember)
}

func (h *Handler) listWorkspaces(c *gin.Context) {
user, ok := auth.CurrentUser(c)
if !ok {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
return
}

_, _ = h.store.EnsurePersonalWorkspace(c.Request.Context(), user.ID, user.Username)

isAdmin := user.Role == "admin"

workspaces, err := h.store.ListUserWorkspaces(c.Request.Context(), user.ID, isAdmin)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить рабочие области"})
return
}

c.JSON(http.StatusOK, workspaces)
}

func (h *Handler) createWorkspace(c *gin.Context) {
user, ok := auth.CurrentUser(c)
if !ok {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
return
}

var input struct {
Name string `json:"name"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

workspace, err := h.store.CreateWorkspace(c.Request.Context(), input.Name, user.ID)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось создать рабочую область"})
return
}

c.JSON(http.StatusCreated, workspace)
}

func (h *Handler) listMembers(c *gin.Context) {
user, workspaceID, ok := h.requireWorkspaceAccess(c)
if !ok {
return
}

isAdmin := user.Role == "admin"

canAccess, err := h.store.UserCanAccessWorkspace(c.Request.Context(), workspaceID, user.ID, isAdmin)
if err != nil || !canAccess {
c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к рабочей области"})
return
}

members, err := h.store.ListMembers(c.Request.Context(), workspaceID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить участников"})
return
}

c.JSON(http.StatusOK, members)
}

func (h *Handler) addMember(c *gin.Context) {
user, workspaceID, ok := h.requireWorkspaceAccess(c)
if !ok {
return
}

isAdmin := user.Role == "admin"

canManage, err := h.store.UserCanManageWorkspace(c.Request.Context(), workspaceID, user.ID, isAdmin)
if err != nil || !canManage {
c.JSON(http.StatusForbidden, gin.H{"error": "Нет прав на управление рабочей областью"})
return
}

var input struct {
Username string `json:"username"`
Role     string `json:"role"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

targetUser, _, err := h.authStore.GetUserByUsername(c.Request.Context(), authUsername(input.Username))
if err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
return
}

if err := h.store.AddMember(c.Request.Context(), workspaceID, targetUser.ID, input.Role); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось добавить пользователя"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Пользователь добавлен"})
}

func (h *Handler) removeMember(c *gin.Context) {
user, workspaceID, ok := h.requireWorkspaceAccess(c)
if !ok {
return
}

isAdmin := user.Role == "admin"

canManage, err := h.store.UserCanManageWorkspace(c.Request.Context(), workspaceID, user.ID, isAdmin)
if err != nil || !canManage {
c.JSON(http.StatusForbidden, gin.H{"error": "Нет прав на управление рабочей областью"})
return
}

targetUserID, err := strconv.Atoi(c.Param("userID"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
return
}

if targetUserID == user.ID {
c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя удалить самого себя из рабочей области"})
return
}

removed, err := h.store.RemoveMember(c.Request.Context(), workspaceID, targetUserID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить участника"})
return
}

if !removed {
c.JSON(http.StatusNotFound, gin.H{"error": "Участник не найден"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Участник удалён"})
}

func (h *Handler) requireWorkspaceAccess(c *gin.Context) (auth.User, int, bool) {
user, ok := auth.CurrentUser(c)
if !ok {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
return auth.User{}, 0, false
}

workspaceID, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID рабочей области"})
return auth.User{}, 0, false
}

return user, workspaceID, true
}

func authUsername(username string) string {
return normalizeUsernameForAuth(username)
}

func normalizeUsernameForAuth(username string) string {
result := ""

for _, ch := range username {
if ch >= 'A' && ch <= 'Z' {
result += string(ch + 32)
continue
}

if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
result += string(ch)
}
}

return result
}
