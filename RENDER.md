# Render Deploy Notes

Проект задеплоен на Render как Web Service.

Live demo:
https://dev-dashboard-557n.onrender.com

## Web Service

Runtime:
Go

Build Command:
go build -o app ./cmd/server

Start Command:
./app

## Environment Variables

PORT
Задается Render автоматически.

DATABASE_URL
Строка подключения к PostgreSQL базе на Render.

## PostgreSQL

База данных:
dev-dashboard-db

Используется Render PostgreSQL.

Приложение берет DATABASE_URL из Environment Variables.

Если DATABASE_URL не задана, локально используется fallback через data/tasks.json.

## Проверка после деплоя

1. Открыть live demo.
2. Создать задачу.
3. Обновить страницу.
4. Изменить статус задачи.
5. Обновить страницу еще раз.
6. Проверить, что данные сохранились.

## Важное замечание

На бесплатном Render сервис может засыпать при неактивности.
Первое открытие после простоя может занять 30-60 секунд.
