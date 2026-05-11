# Dev Dashboard

Публичный multi-workspace dashboard для управления задачами.

Live demo: https://dev-dashboard-557n.onrender.com
GitHub: https://github.com/maystrenkobr-ops/dev-dashboard

## Скриншот

![Dev Dashboard](docs/screenshot.png)

## Описание

Dev Dashboard — мини task manager / Kanban dashboard на Go + Gin с PostgreSQL, авторизацией, админ-панелью и рабочими областями.

Проект поддерживает личное использование, командные рабочие области и разделение доступа между пользователями.

## Возможности

- регистрация и вход пользователей
- bcrypt-хеширование паролей
- cookie-сессии
- admin-панель управления пользователями
- рабочие области
- личные рабочие области пользователей
- добавление участников в рабочую область
- роли workspace owner и member
- global admin видит и модерирует всё
- Kanban-доска todo / in_progress / done
- создание, редактирование и удаление задач
- drag-and-drop карточек между колонками
- приоритеты low / medium / high
- дедлайны задач
- дата и время создания задач
- московское время Europe/Moscow
- подсветка просроченных и сегодняшних дедлайнов
- поиск по задачам
- фильтр по приоритету
- PostgreSQL-хранилище на Render
- локальный JSON fallback через data/tasks.json

## Роли и доступы

### Global admin

- видит все рабочие области
- видит всех пользователей
- может удалять пользователей
- может модерировать участников любой рабочей области
- может работать с задачами в любой области

### Workspace owner

- управляет своей рабочей областью
- добавляет участников
- удаляет участников
- создаёт и редактирует задачи

### Workspace member

- видит только доступные ему рабочие области
- видит задачи внутри этих областей
- может создавать и редактировать задачи
- не видит список участников
- не может добавлять или удалять участников

## Стек

- Go
- Gin
- PostgreSQL
- pgx / pgxpool
- bcrypt
- HTML
- CSS
- JavaScript
- Render
- GitHub
- Docker

## Основная структура проекта

- cmd/server/main.go
- internal/auth
- internal/tasks
- internal/workspaces
- web
- docs/screenshot.png
- RENDER.md
- Dockerfile

## Основные API

Auth:
- POST /api/register
- POST /api/login
- POST /api/logout
- GET /api/me

Admin:
- GET /admin/users
- GET /api/admin/users
- DELETE /api/admin/users/:id

Workspaces:
- GET /api/workspaces
- POST /api/workspaces
- GET /api/workspaces/:id/members
- POST /api/workspaces/:id/members
- DELETE /api/workspaces/:id/members/:userID

Tasks:
- GET /tasks?workspace_id=:id
- POST /tasks?workspace_id=:id
- PATCH /tasks/:id/title?workspace_id=:id
- PATCH /tasks/:id/status?workspace_id=:id
- PATCH /tasks/:id/priority?workspace_id=:id
- PATCH /tasks/:id/deadline?workspace_id=:id
- DELETE /tasks/:id?workspace_id=:id

## Локальный запуск

go run .\cmd\server\main.go

После запуска открыть:

http://localhost:8080

## Переменные окружения

- DATABASE_URL
- ADMIN_USERNAME
- ADMIN_PASSWORD
- SESSION_SECRET
- PORT

## Деплой

Проект задеплоен на Render как Web Service.

Build Command:
go build -o app ./cmd/server

Start Command:
./app

## Статус проекта

Готовый публичный portfolio-проект с авторизацией, PostgreSQL, рабочими областями, ролями, Kanban UI, drag-and-drop, дедлайнами, поиском, фильтрами и деплоем на Render.
