package auth

import (
"crypto/hmac"
"crypto/sha256"
"encoding/base64"
"encoding/json"
"net/http"
"strconv"
"strings"
"time"

"github.com/gin-gonic/gin"
)

const sessionCookieName = "dev_dashboard_session"
const userContextKey = "auth_user"

type SessionManager struct {
secret []byte
}

type sessionPayload struct {
UserID   int    `json:"user_id"`
Username string `json:"username"`
Role     string `json:"role"`
Exp      int64  `json:"exp"`
}

func NewSessionManager(secret string) *SessionManager {
if strings.TrimSpace(secret) == "" {
secret = "local-dev-secret-change-me"
}

return &SessionManager{
secret: []byte(secret),
}
}

func (m *SessionManager) SetSession(c *gin.Context, user User) error {
token, err := m.createToken(user)
if err != nil {
return err
}

http.SetCookie(c.Writer, &http.Cookie{
Name:     sessionCookieName,
Value:    token,
Path:     "/",
MaxAge:   60 * 60 * 24 * 7,
HttpOnly: true,
Secure:   isHTTPS(c),
SameSite: http.SameSiteLaxMode,
})

return nil
}

func (m *SessionManager) ClearSession(c *gin.Context) {
http.SetCookie(c.Writer, &http.Cookie{
Name:     sessionCookieName,
Value:    "",
Path:     "/",
MaxAge:   -1,
HttpOnly: true,
Secure:   isHTTPS(c),
SameSite: http.SameSiteLaxMode,
})
}

func (m *SessionManager) RequireAuth(store Store) gin.HandlerFunc {
return func(c *gin.Context) {
payload, ok := m.readSession(c)
if !ok {
respondUnauthorized(c)
return
}

user, found, err := store.GetUserByID(c.Request.Context(), payload.UserID)
if err != nil || !found {
m.ClearSession(c)
respondUnauthorized(c)
return
}

c.Set(userContextKey, user)
c.Next()
}
}

func (m *SessionManager) RequireAdmin() gin.HandlerFunc {
return func(c *gin.Context) {
value, ok := c.Get(userContextKey)
if !ok {
respondUnauthorized(c)
return
}

user, ok := value.(User)
if !ok || user.Role != "admin" {
c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
c.Abort()
return
}

c.Next()
}
}

func CurrentUser(c *gin.Context) (User, bool) {
value, ok := c.Get(userContextKey)
if !ok {
return User{}, false
}

user, ok := value.(User)
return user, ok
}

func (m *SessionManager) createToken(user User) (string, error) {
payload := sessionPayload{
UserID:   user.ID,
Username: user.Username,
Role:     user.Role,
Exp:      time.Now().Add(7 * 24 * time.Hour).Unix(),
}

data, err := json.Marshal(payload)
if err != nil {
return "", err
}

encodedPayload := base64.RawURLEncoding.EncodeToString(data)
signature := m.sign(encodedPayload)

return encodedPayload + "." + signature, nil
}

func (m *SessionManager) readSession(c *gin.Context) (sessionPayload, bool) {
cookie, err := c.Cookie(sessionCookieName)
if err != nil {
return sessionPayload{}, false
}

parts := strings.Split(cookie, ".")
if len(parts) != 2 {
return sessionPayload{}, false
}

expectedSignature := m.sign(parts[0])
if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
return sessionPayload{}, false
}

data, err := base64.RawURLEncoding.DecodeString(parts[0])
if err != nil {
return sessionPayload{}, false
}

var payload sessionPayload

if err := json.Unmarshal(data, &payload); err != nil {
return sessionPayload{}, false
}

if payload.Exp < time.Now().Unix() {
return sessionPayload{}, false
}

return payload, true
}

func (m *SessionManager) sign(payload string) string {
mac := hmac.New(sha256.New, m.secret)
mac.Write([]byte(payload))

return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func isHTTPS(c *gin.Context) bool {
return c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
}

func respondUnauthorized(c *gin.Context) {
if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/tasks") {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется вход"})
c.Abort()
return
}

c.Redirect(http.StatusFound, "/login")
c.Abort()
}

func UserIDFromParam(c *gin.Context) (int, bool) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
return 0, false
}

return id, true
}
