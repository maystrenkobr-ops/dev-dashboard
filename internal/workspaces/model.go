package workspaces

type Workspace struct {
ID        int    `json:"id"`
Name      string `json:"name"`
OwnerID   int    `json:"owner_id"`
CreatedAt string `json:"created_at"`
}

type Member struct {
WorkspaceID int    `json:"workspace_id"`
UserID      int    `json:"user_id"`
Username    string `json:"username"`
Role        string `json:"role"`
CreatedAt   string `json:"created_at"`
}
