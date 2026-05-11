package workspaces

import (
"context"
"errors"

"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
db *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
pool, err := pgxpool.New(ctx, databaseURL)
if err != nil {
return nil, err
}

store := &PostgresStore{db: pool}

if err := store.createTables(ctx); err != nil {
pool.Close()
return nil, err
}

return store, nil
}

func (s *PostgresStore) Close() {
s.db.Close()
}

func (s *PostgresStore) CreateWorkspace(ctx context.Context, name string, ownerID int) (Workspace, error) {
name = normalizeName(name)
if name == "" {
return Workspace{}, errors.New("workspace name is required")
}

var workspace Workspace

err := s.db.QueryRow(ctx, `
INSERT INTO workspaces (name, owner_id)
VALUES ($1, $2)
RETURNING id, name, owner_id, to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI');
`, name, ownerID).Scan(&workspace.ID, &workspace.Name, &workspace.OwnerID, &workspace.CreatedAt)
if err != nil {
return Workspace{}, err
}

if err := s.AddMember(ctx, workspace.ID, ownerID, "owner"); err != nil {
return Workspace{}, err
}

return workspace, nil
}

func (s *PostgresStore) GetWorkspace(ctx context.Context, workspaceID int) (Workspace, bool, error) {
var workspace Workspace

err := s.db.QueryRow(ctx, `
SELECT id, name, owner_id, to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM workspaces
WHERE id = $1;
`, workspaceID).Scan(&workspace.ID, &workspace.Name, &workspace.OwnerID, &workspace.CreatedAt)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return Workspace{}, false, nil
}

return Workspace{}, false, err
}

return workspace, true, nil
}

func (s *PostgresStore) ListUserWorkspaces(ctx context.Context, userID int, isAdmin bool) ([]Workspace, error) {
var rows pgx.Rows
var err error

if isAdmin {
rows, err = s.db.Query(ctx, `
SELECT id, name, owner_id, to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM workspaces
ORDER BY id;
`)
} else {
rows, err = s.db.Query(ctx, `
SELECT w.id, w.name, w.owner_id, to_char(w.created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM workspaces w
INNER JOIN workspace_members wm ON wm.workspace_id = w.id
WHERE wm.user_id = $1
ORDER BY w.id;
`, userID)
}

if err != nil {
return nil, err
}
defer rows.Close()

result := []Workspace{}

for rows.Next() {
var workspace Workspace

if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.OwnerID, &workspace.CreatedAt); err != nil {
return nil, err
}

result = append(result, workspace)
}

return result, rows.Err()
}

func (s *PostgresStore) ListMembers(ctx context.Context, workspaceID int) ([]Member, error) {
rows, err := s.db.Query(ctx, `
SELECT wm.workspace_id, wm.user_id, u.username, wm.role, to_char(wm.created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM workspace_members wm
INNER JOIN app_users u ON u.id = wm.user_id
WHERE wm.workspace_id = $1
ORDER BY wm.user_id;
`, workspaceID)
if err != nil {
return nil, err
}
defer rows.Close()

result := []Member{}

for rows.Next() {
var member Member

if err := rows.Scan(&member.WorkspaceID, &member.UserID, &member.Username, &member.Role, &member.CreatedAt); err != nil {
return nil, err
}

result = append(result, member)
}

return result, rows.Err()
}

func (s *PostgresStore) AddMember(ctx context.Context, workspaceID int, userID int, role string) error {
_, err := s.db.Exec(ctx, `
INSERT INTO workspace_members (workspace_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (workspace_id, user_id)
DO UPDATE SET role = EXCLUDED.role;
`, workspaceID, userID, normalizeRole(role))

return err
}

func (s *PostgresStore) RemoveMember(ctx context.Context, workspaceID int, userID int) (bool, error) {
result, err := s.db.Exec(ctx, `
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;
`, workspaceID, userID)
if err != nil {
return false, err
}

return result.RowsAffected() > 0, nil
}

func (s *PostgresStore) UserCanAccessWorkspace(ctx context.Context, workspaceID int, userID int, isAdmin bool) (bool, error) {
if isAdmin {
return true, nil
}

var exists bool

err := s.db.QueryRow(ctx, `
SELECT EXISTS (
SELECT 1
FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2
);
`, workspaceID, userID).Scan(&exists)

return exists, err
}

func (s *PostgresStore) UserCanManageWorkspace(ctx context.Context, workspaceID int, userID int, isAdmin bool) (bool, error) {
if isAdmin {
return true, nil
}

var exists bool

err := s.db.QueryRow(ctx, `
SELECT EXISTS (
SELECT 1
FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2 AND role = 'owner'
);
`, workspaceID, userID).Scan(&exists)

return exists, err
}

func (s *PostgresStore) EnsurePersonalWorkspace(ctx context.Context, userID int, username string) (Workspace, error) {
name := "Личная область: " + username

var workspace Workspace

err := s.db.QueryRow(ctx, `
SELECT id, name, owner_id, to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM workspaces
WHERE owner_id = $1 AND name = $2;
`, userID, name).Scan(&workspace.ID, &workspace.Name, &workspace.OwnerID, &workspace.CreatedAt)

if err == nil {
_ = s.AddMember(ctx, workspace.ID, userID, "owner")
return workspace, nil
}

if !errors.Is(err, pgx.ErrNoRows) {
return Workspace{}, err
}

return s.CreateWorkspace(ctx, name, userID)
}

func (s *PostgresStore) EnsureDefaultWorkspace(ctx context.Context, adminUserID int) (Workspace, error) {
var workspace Workspace

err := s.db.QueryRow(ctx, `
SELECT id, name, owner_id, to_char(created_at AT TIME ZONE 'Europe/Moscow', 'YYYY-MM-DD HH24:MI')
FROM workspaces
WHERE name = 'Общая доска'
ORDER BY id
LIMIT 1;
`).Scan(&workspace.ID, &workspace.Name, &workspace.OwnerID, &workspace.CreatedAt)

if err == nil {
if adminUserID > 0 {
_ = s.AddMember(ctx, workspace.ID, adminUserID, "owner")
}

return workspace, nil
}

if !errors.Is(err, pgx.ErrNoRows) {
return Workspace{}, err
}

return s.CreateWorkspace(ctx, "Общая доска", adminUserID)
}

func (s *PostgresStore) createTables(ctx context.Context) error {
_, err := s.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workspaces (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
owner_id INTEGER NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workspace_members (
workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
user_id INTEGER NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
role TEXT NOT NULL DEFAULT 'member',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
PRIMARY KEY (workspace_id, user_id)
);
`)

return err
}
