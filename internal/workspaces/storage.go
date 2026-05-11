package workspaces

import (
"context"
"strings"
)

type Store interface {
Close()
CreateWorkspace(ctx context.Context, name string, ownerID int) (Workspace, error)
GetWorkspace(ctx context.Context, workspaceID int) (Workspace, bool, error)
ListUserWorkspaces(ctx context.Context, userID int, isAdmin bool) ([]Workspace, error)
ListMembers(ctx context.Context, workspaceID int) ([]Member, error)
AddMember(ctx context.Context, workspaceID int, userID int, role string) error
RemoveMember(ctx context.Context, workspaceID int, userID int) (bool, error)
UserCanAccessWorkspace(ctx context.Context, workspaceID int, userID int, isAdmin bool) (bool, error)
UserCanManageWorkspace(ctx context.Context, workspaceID int, userID int, isAdmin bool) (bool, error)
EnsurePersonalWorkspace(ctx context.Context, userID int, username string) (Workspace, error)
EnsureDefaultWorkspace(ctx context.Context, adminUserID int) (Workspace, error)
}

func NewStorage(ctx context.Context, databaseURL string) (Store, error) {
if databaseURL != "" {
return NewPostgresStore(ctx, databaseURL)
}

return NewMemoryStore(), nil
}

func normalizeName(name string) string {
return strings.TrimSpace(name)
}

func normalizeRole(role string) string {
role = strings.ToLower(strings.TrimSpace(role))

if role != "owner" && role != "member" {
return "member"
}

return role
}
