package auth

import (
"context"
"strings"
)

type Store interface {
Close()
CreateUser(ctx context.Context, username string, password string) (User, error)
GetUserByUsername(ctx context.Context, username string) (User, string, error)
GetUserByID(ctx context.Context, id int) (User, bool, error)
ListUsers(ctx context.Context) ([]User, error)
DeleteUser(ctx context.Context, id int) (bool, error)
}

func NewStorage(ctx context.Context, databaseURL string, adminUsername string, adminPassword string) (Store, error) {
adminUsername = normalizeUsername(adminUsername)

if databaseURL != "" {
store, err := NewPostgresStore(ctx, databaseURL)
if err != nil {
return nil, err
}

if adminUsername != "" && adminPassword != "" {
if err := store.SeedAdmin(ctx, adminUsername, adminPassword); err != nil {
store.Close()
return nil, err
}
}

return store, nil
}

store := NewMemoryStore()

if adminUsername != "" && adminPassword != "" {
if err := store.SeedAdmin(ctx, adminUsername, adminPassword); err != nil {
return nil, err
}
}

return store, nil
}

func normalizeUsername(username string) string {
return strings.ToLower(strings.TrimSpace(username))
}
