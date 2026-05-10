package auth

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

if err := store.createUsersTable(ctx); err != nil {
pool.Close()
return nil, err
}

return store, nil
}

func (s *PostgresStore) Close() {
s.db.Close()
}

func (s *PostgresStore) SeedAdmin(ctx context.Context, username string, password string) error {
hash, err := hashPassword(password)
if err != nil {
return err
}

_, err = s.db.Exec(ctx, `
INSERT INTO app_users (username, password_hash, role)
VALUES ($1, $2, 'admin')
ON CONFLICT (username)
DO UPDATE SET password_hash = EXCLUDED.password_hash, role = 'admin';
`, username, hash)

return err
}

func (s *PostgresStore) CreateUser(ctx context.Context, username string, password string) (User, error) {
hash, err := hashPassword(password)
if err != nil {
return User{}, err
}

var user User

err = s.db.QueryRow(ctx, `
INSERT INTO app_users (username, password_hash, role)
VALUES ($1, $2, 'user')
RETURNING id, username, role, created_at::text;
`, username, hash).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)

return user, err
}

func (s *PostgresStore) GetUserByUsername(ctx context.Context, username string) (User, string, error) {
var user User
var passwordHash string

err := s.db.QueryRow(ctx, `
SELECT id, username, role, created_at::text, password_hash
FROM app_users
WHERE username = $1;
`, username).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt, &passwordHash)

return user, passwordHash, err
}

func (s *PostgresStore) GetUserByID(ctx context.Context, id int) (User, bool, error) {
var user User

err := s.db.QueryRow(ctx, `
SELECT id, username, role, created_at::text
FROM app_users
WHERE id = $1;
`, id).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)

if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return User{}, false, nil
}

return User{}, false, err
}

return user, true, nil
}

func (s *PostgresStore) ListUsers(ctx context.Context) ([]User, error) {
rows, err := s.db.Query(ctx, `
SELECT id, username, role, created_at::text
FROM app_users
ORDER BY id;
`)
if err != nil {
return nil, err
}
defer rows.Close()

result := []User{}

for rows.Next() {
var user User

if err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt); err != nil {
return nil, err
}

result = append(result, user)
}

return result, rows.Err()
}

func (s *PostgresStore) DeleteUser(ctx context.Context, id int) (bool, error) {
result, err := s.db.Exec(ctx, `
DELETE FROM app_users
WHERE id = $1;
`, id)
if err != nil {
return false, err
}

return result.RowsAffected() > 0, nil
}

func (s *PostgresStore) createUsersTable(ctx context.Context) error {
_, err := s.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS app_users (
id SERIAL PRIMARY KEY,
username TEXT NOT NULL UNIQUE,
password_hash TEXT NOT NULL,
role TEXT NOT NULL DEFAULT 'user',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`)

return err
}
