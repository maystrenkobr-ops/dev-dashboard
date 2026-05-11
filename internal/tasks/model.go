package tasks

type Task struct {
ID          int    `json:"id"`
WorkspaceID int    `json:"workspace_id"`
CreatedBy   int    `json:"created_by"`
Title       string `json:"title"`
Status      string `json:"status"`
Priority    string `json:"priority"`
Deadline    string `json:"deadline"`
CreatedAt   string `json:"created_at"`
}
