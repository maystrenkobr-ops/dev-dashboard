package auth

import (
"context"
"errors"
"sync"
"time"

"golang.org/x/crypto/bcrypt"
)

type memoryUserRecord struct {
User
PasswordHash string
}

type MemoryStore struct {
mu     sync.Mutex
users  map[int]memoryUserRecord
nextID int
}

func NewMemoryStore() *MemoryStore {
return &MemoryStore{
users:  map[int]memoryUserRecord{},
nextID: 1,
}
}

func (s *MemoryStore) Close() {}

func (s *MemoryStore) SeedAdmin(ctx context.Context, username string, password string) error {
s.mu.Lock()
defer s.mu.Unlock()

hash, err := hashPassword(password)
if err != nil {
return err
}

for id, record := range s.users {
if record.Username == username {
record.PasswordHash = hash
record.Role = "admin"
s.users[id] = record
return nil
}
}

user := User{
ID:        s.nextID,
Username:  username,
Role:      "admin",
CreatedAt: time.Now().Format(time.RFC3339),
}

s.users[user.ID] = memoryUserRecord{
User:         user,
PasswordHash: hash,
}

s.nextID++

return nil
}

func (s *MemoryStore) CreateUser(ctx context.Context, username string, password string) (User, error) {
s.mu.Lock()
defer s.mu.Unlock()

for _, record := range s.users {
if record.Username == username {
return User{}, errors.New("username already exists")
}
}

hash, err := hashPassword(password)
if err != nil {
return User{}, err
}

user := User{
ID:        s.nextID,
Username:  username,
Role:      "user",
CreatedAt: time.Now().Format(time.RFC3339),
}

s.users[user.ID] = memoryUserRecord{
User:         user,
PasswordHash: hash,
}

s.nextID++

return user, nil
}

func (s *MemoryStore) GetUserByUsername(ctx context.Context, username string) (User, string, error) {
s.mu.Lock()
defer s.mu.Unlock()

for _, record := range s.users {
if record.Username == username {
return record.User, record.PasswordHash, nil
}
}

return User{}, "", errors.New("user not found")
}

func (s *MemoryStore) GetUserByID(ctx context.Context, id int) (User, bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

record, ok := s.users[id]
if !ok {
return User{}, false, nil
}

return record.User, true, nil
}

func (s *MemoryStore) ListUsers(ctx context.Context) ([]User, error) {
s.mu.Lock()
defer s.mu.Unlock()

result := make([]User, 0, len(s.users))

for _, record := range s.users {
result = append(result, record.User)
}

return result, nil
}

func (s *MemoryStore) DeleteUser(ctx context.Context, id int) (bool, error) {
s.mu.Lock()
defer s.mu.Unlock()

if _, ok := s.users[id]; !ok {
return false, nil
}

delete(s.users, id)

return true, nil
}

func hashPassword(password string) (string, error) {
data, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
if err != nil {
return "", err
}

return string(data), nil
}
