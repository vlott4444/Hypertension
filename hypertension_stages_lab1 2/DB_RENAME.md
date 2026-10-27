# Переименование таблиц в существующей PostgreSQL БД

Если БД уже содержит данные, до запуска новой версии приложения выполните SQL в PostgreSQL:

```sql
BEGIN;
ALTER TABLE IF EXISTS hypertension_services RENAME TO hypertension_stages;
ALTER TABLE IF EXISTS hypertension_likes RENAME COLUMN hypertension_service_id TO hypertension_stage_id;
COMMIT;
```

Сделайте резервную копию БД перед запуском. PostgreSQL сохраняет существующие внешние ключи при переименовании таблицы и столбца. `AutoMigrate` не является миграцией-переименованием и сам по себе может создать новые таблицы/колонки вместо старых. Если в БД уже существуют таблицы с новыми названиями, перед применением SQL потребуется сверить схему и данные.
