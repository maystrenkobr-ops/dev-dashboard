package workspaces

import (
"context"
"errors"
"sync"
"time"
)

type MemoryStore struct {
mu          sync.Mutex
workspaces map[int]Workspace
members    map[int]map[int]Member
nextID      int
}

func NewMemoryStore() *MemoryStore {
return &MemoryStore{
workspaces: map[int]Workspace{},
members:    map[int]map[int]Member{},
nextID:      1,
}
}

func (s *MemoryStore) Close() {}

func (s *MemoryStore) CreateWorkspace(ctx context.Context, name string, ownerID int) (Workspace, error) {
s.mu.Lock()
defer s.mu.Unlock()

name = normalizeName(name)
if name == "" {
return Workspace{}, errors.New("workspace name is required")
}

workspace := Workspace{
ID:        s.nextID,
Name:      name,
OwnerID:   ownerID,
CreatedAt: time.Now().Format("2006-01-02 15:04"),
}

s.nextID++
s.workspaces[workspace.ID] = workspace

if s.members[workspace.ID] == nil {
s.members[workspace.ID] = map[int]Member{}
}

s.members[workspace.ID][ownerID] = Member{
WorkspaceID: workspace.ID,
UserID:      ownerID,
Username:    "",
Role:        "owner",
CreatedAt:   time.Now().Format("2006-01-02 15:04"),
}

return workspace, nil
}

func (s *MemoryStore) GetWorkspace(ctx context.Context, workspaceID int) (Workspace, bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

workspace, ok := s.workspaces[workspaceID]
return workspace, ok, nil
}

func (s *MemoryStore) ListUserWorkspaces(ctx context.Context, userID int, isAdmin bool) ([]Workspace, error) {
s.mu.Lock()
defer s.mu.Unlock()

result := []Workspace{}

for id, workspace := range s.workspaces {
if isAdmin {
result = append(result, workspace)
continue
}

if s.members[id] != nil {
if _, ok := s.members[id][userID]; ok {
result = append(result, workspace)
}
}
}

return result, nil
}

func (s *MemoryStore) ListMembers(ctx context.Context, workspaceID int) ([]Member, error) {
s.mu.Lock()
defer s.mu.Unlock()

result := []Member{}

for _, member := range s.members[workspaceID] {
result = append(result, member)
}

return result, nil
}

func (s *MemoryStore) AddMember(ctx context.Context, workspaceID int, userID int, role string) error {
s.mu.Lock()
defer s.mu.Unlock()

if _, ok := s.workspaces[workspaceID]; !ok {
return errors.New("workspace not found")
}

if s.members[workspaceID] == nil {
s.members[workspaceID] = map[int]Member{}
}

s.members[workspaceID][userID] = Member{
WorkspaceID: workspaceID,
UserID:      userID,
Username:    "",
Role:        normalizeRole(role),
CreatedAt:   time.Now().Format("2006-01-02 15:04"),
}

return nil
}

func (s *MemoryStore) RemoveMember(ctx context.Context, workspaceID int, userID int) (bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

if s.members[workspaceID] == nil {
return false, nil
}

if _, ok := s.members[workspaceID][userID]; !ok {
return false, nil
}

delete(s.members[workspaceID], userID)

return true, nil
}

func (s *MemoryStore) UserCanAccessWorkspace(ctx context.Context, workspaceID int, userID int, isAdmin bool) (bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

if isAdmin {
return true, nil
}

if s.members[workspaceID] == nil {
return false, nil
}

_, ok := s.members[workspaceID][userID]
return ok, nil
}

func (s *MemoryStore) UserCanManageWorkspace(ctx context.Context, workspaceID int, userID int, isAdmin bool) (bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

if isAdmin {
return true, nil
}

member, ok := s.members[workspaceID][userID]
if !ok {
return false, nil
}

return member.Role == "owner", nil
}

func (s *MemoryStore) EnsurePersonalWorkspace(ctx context.Context, userID int, username string) (Workspace, error) {
s.mu.Lock()
defer s.mu.Unlock()

name := "Личная область: " + username

for _, workspace := range s.workspaces {
if workspace.OwnerID == userID && workspace.Name == name {
return workspace, nil
}
}

workspace := Workspace{
ID:        s.nextID,
Name:      name,
OwnerID:   userID,
CreatedAt: time.Now().Format("2006-01-02 15:04"),
}

s.nextID++
s.workspaces[workspace.ID] = workspace

if s.members[workspace.ID] == nil {
s.members[workspace.ID] = map[int]Member{}
}

s.members[workspace.ID][userID] = Member{
WorkspaceID: workspace.ID,
UserID:      userID,
Username:    username,
Role:        "owner",
CreatedAt:   time.Now().Format("2006-01-02 15:04"),
}

return workspace, nil
}

func (s *MemoryStore) EnsureDefaultWorkspace(ctx context.Context, adminUserID int) (Workspace, error) {
s.mu.Lock()
defer s.mu.Unlock()

for _, workspace := range s.workspaces {
if workspace.Name == "Общая доска" {
return workspace, nil
}
}

workspace := Workspace{
ID:        s.nextID,
Name:      "Общая доска",
OwnerID:   adminUserID,
CreatedAt: time.Now().Format("2006-01-02 15:04"),
}

s.nextID++
s.workspaces[workspace.ID] = workspace

if s.members[workspace.ID] == nil {
s.members[workspace.ID] = map[int]Member{}
}

if adminUserID > 0 {
s.members[workspace.ID][adminUserID] = Member{
WorkspaceID: workspace.ID,
UserID:      adminUserID,
Username:    "",
Role:        "owner",
CreatedAt:   time.Now().Format("2006-01-02 15:04"),
}
}

return workspace, nil
}
