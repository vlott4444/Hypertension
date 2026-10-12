# Hypertension Stages — REST API

Веб-сервис для работы со стадиями гипертонии на Go + Gin + GORM + PostgreSQL + Minio.

## Стек

- **Go** — язык
- **Gin** — HTTP-роутинг
- **GORM** — ORM для PostgreSQL
- **PostgreSQL** — БД
- **MinIO** — S3-хранилище для медиа опубликованных карточек
- **Локальный сервер** (`resources/media`) — фото и видео черновика по умолчанию
- **Insomnia / Postman** — тестирование API

## Запуск

```bash
go run cmd/migrate/main.go   # миграции
go run cmd/seed/main.go      # заполнение БД
go run cmd/server/main.go    # запуск сервера на :8080


## REST API

Основные маршруты: `GET /api/hypertension_stages`, `POST /api/hypertension_stages`,
`PUT /api/hypertension_stages/:id/publish`, `DELETE /api/hypertension_stages/:id`,
`POST /api/hypertension_stages/:id/like`.
Ответ списка использует ключ `hypertension_stages`, ответ отдельной карточки — `hypertension_stage`.

## Медиа черновика

Стандартное изображение `/resources/media/draft_stage1.png` и видео
`/resources/media/draft_stage1.mp4` раздаются непосредственно Gin-сервером,
без обращения к MinIO. Путь к медиа: `router.Static("/resources", "./resources")`.
Это не меняет текущую заглушку `saveUploadedFile`: загрузка пользовательских
медиафайлов в MinIO в исходном проекте ещё не реализована.
