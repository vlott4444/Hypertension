# Лабораторная № 4 — Hypertension Stages

## Запуск

1. Установить Go версии, соответствующей `go.mod` (указана 1.27), Docker/Compose.
2. `docker compose up -d` запускает PostgreSQL, Adminer, MinIO и Redis.
3. Создать `.env` на основе текущих параметров БД проекта и `.env.lab4.example`; задать **случайный `JWT_SECRET` длиной >= 32 символов**, `REDIS_ADDR=localhost:6379`, MINIO_*. Не коммитить `.env`.
4. `go run ./cmd/migrate` (после резервного копирования БД), затем `go run ./cmd/server`.
5. Открыть http://localhost:8080/swagger (Swagger UI использует ресурсы CDN; нужен интернет; схема доступна локально по `/swagger/openapi.json`).
6. Зарегистрировать нового пользователя через `POST /api/auth/register` и войти `POST /api/auth/login`. Старые пользователи из seed с паролями открытым текстом не смогут войти — для них нужно отдельно выполнить безопасный сброс пароля.
7. Нажать `Authorize` и вставить JWT без слова `Bearer` (Swagger подставляет схему автоматически).

## Методы и безопасность

| Метод | URL | Гость | Создатель |
|---|---|---|---|
| POST | /api/auth/register | Да | Да |
| POST | /api/auth/login | Да | Да |
| POST | /api/auth/logout | Нет | Да |
| GET | /api/hypertension_stages?filter= | Да | Да |
| GET | /api/feed | Да | Да |
| GET | /api/feed/{id}?next=true | Да | Да |
| GET | /api/draft | Нет | Да |
| POST | /api/hypertension_stages | Нет | Да |
| PUT | /api/hypertension_stages/{id}/publish | Нет | Да, только своё |
| DELETE | /api/hypertension_stages/{id} | Нет | Да, только своё |
| POST | /api/hypertension_stages/{id}/like | Нет | Да |

JWT HS256 действует 1 час. Logout ставит `jwt:blacklist:<jti>` в Redis на оставшееся время действия токена. При недоступном Redis защищённые операции блокируются (503). Пароли новых пользователей хранятся как bcrypt.

В ЛР 4 Swagger проверяет REST маршруты `/api/*`; существующие HTML/SSR маршруты `/draft`, `/stages` и т. п. **остались старыми и не защищены JWT**. Перед реальным развёртыванием закрыть HTML-операции или переключить их на авторизованный API.

Загрузка медиа из REST `POST /api/hypertension_stages` использует S3-compatible `PUT` в MinIO (PNG до 8 МБ и MP4 до 20 МБ). Локальные стандартные фото и видео остаются в `resources/media`. Публичное чтение файлов требует подходящей политики bucket.

В проекте не реализована полноценная транзакция между загрузкой файлов и созданием записи в БД: при ошибке БД после успешного upload могут остаться лишние объекты MinIO. Для production потребуется cleanup/transaction orchestration.

## Демонстрация по скриншотам

1–3. Инкогнито, Swagger: логин, JWT, список опубликованных стадий.
4–5. GET списка с фильтром в Postman без Authorization — 200 и тот же результат.
6. PUT publish без Authorization — 401.
7. PUT publish c `Authorization: Bearer <JWT>` для своего черновика — 200.
8. Redis CLI: `docker exec -it hypertension_redis redis-cli KEYS 'jwt:blacklist:*'` — до logout пусто; после logout виден JTI в blacklist. Для JWT **в Redis не хранится пользовательская сессия**: пользователя определяют из подписанного `sub`. При необходимости продемонстрировать `GET jwt:blacklist:<jti>`.
9. StarUML: Sequence Diagram; у сообщения выбрать `signature` из операций диаграммы классов (вручную в StarUML). Сценарий приведён в `docs/sequence_lab4.puml`.

## Границы проверки

Модуль Go в исходнике требует Go 1.27. В этой рабочей среде установлен 1.23.2, поэтому сборка всего сервера и интеграционные проверки PostgreSQL/Redis/MinIO не выполнены. Изолированные пакеты `auth` и `media` проверены `go test` в режиме стандартной библиотеки.

**Миграция:** для поля `users.username` добавлен уникальный индекс. Перед `AutoMigrate` проверьте отсутствие повторяющихся `username` в существующей БД.
