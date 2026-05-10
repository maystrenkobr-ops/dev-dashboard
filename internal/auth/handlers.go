package auth

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"
"golang.org/x/crypto/bcrypt"
)

func RegisterRoutes(router *gin.Engine, store Store, sessions *SessionManager) {
router.GET("/login", func(c *gin.Context) {
c.File("web/login.html")
})

router.GET("/register", func(c *gin.Context) {
c.File("web/register.html")
})

router.POST("/api/register", func(c *gin.Context) {
var input struct {
Username string `json:"username"`
Password string `json:"password"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

username := normalizeUsername(input.Username)
password := strings.TrimSpace(input.Password)

if len(username) < 3 {
c.JSON(http.StatusBadRequest, gin.H{"error": "Логин должен быть минимум 3 символа"})
return
}

if len(password) < 6 {
c.JSON(http.StatusBadRequest, gin.H{"error": "Пароль должен быть минимум 6 символов"})
return
}

user, err := store.CreateUser(c.Request.Context(), username, password)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Такой логин уже занят"})
return
}

if err := sessions.SetSession(c, user); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
return
}

c.JSON(http.StatusCreated, user)
})

router.POST("/api/login", func(c *gin.Context) {
var input struct {
Username string `json:"username"`
Password string `json:"password"`
}

if err := c.ShouldBindJSON(&input); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный JSON"})
return
}

username := normalizeUsername(input.Username)
password := strings.TrimSpace(input.Password)

user, passwordHash, err := store.GetUserByUsername(c.Request.Context(), username)
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
return
}

if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
return
}

if err := sessions.SetSession(c, user); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
return
}

c.JSON(http.StatusOK, user)
})

router.POST("/api/logout", sessions.RequireAuth(store), func(c *gin.Context) {
sessions.ClearSession(c)
c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен"})
})

router.GET("/api/me", sessions.RequireAuth(store), func(c *gin.Context) {
user, ok := CurrentUser(c)
if !ok {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
return
}

c.JSON(http.StatusOK, user)
})

router.GET("/admin/users", sessions.RequireAuth(store), sessions.RequireAdmin(), func(c *gin.Context) {
c.File("web/admin.html")
})

adminAPI := router.Group("/api/admin")
adminAPI.Use(sessions.RequireAuth(store), sessions.RequireAdmin())

adminAPI.GET("/users", func(c *gin.Context) {
users, err := store.ListUsers(c.Request.Context())
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить пользователей"})
return
}

c.JSON(http.StatusOK, users)
})

adminAPI.DELETE("/users/:id", func(c *gin.Context) {
id, ok := UserIDFromParam(c)
if !ok {
return
}

currentUser, ok := CurrentUser(c)
if !ok {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
return
}

if currentUser.ID == id {
c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя удалить самого себя"})
return
}

deleted, err := store.DeleteUser(c.Request.Context(), id)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить пользователя"})
return
}

if !deleted {
c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Пользователь удалён"})
})
}
