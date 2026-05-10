package auth

type User struct {
ID        int    `json:"id"`
Username  string `json:"username"`
Role      string `json:"role"`
CreatedAt string `json:"created_at"`
}
