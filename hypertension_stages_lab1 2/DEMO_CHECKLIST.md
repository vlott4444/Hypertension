# Что показывать преподавателю: скриншоты 1–19

## 1–7. Тема + Figma + приложение

1. Условие варианта: стадии гипертонической болезни.
2. Figma: страница **лента**.
3. Приложение: `http://localhost:8080/feed/` — тот же мобильный стиль.
4. Figma: страница **добавить**.
5. Приложение: `http://localhost:8080/draft`.
6. Figma: страница **плитка**.
7. Приложение: `http://localhost:8080/stages`.

В макете ключевые цвета перенесены в CSS как `#B6001A`, `#FBF2F3`, `#333333`; сохранены большие скругления и нижняя pill-навигация.

## 8–9. Источник стилистики + CSS

8. Открыть популярный ресурс, который использовался при создании вашего Figma-макета. Его URL в присланных материалах отсутствует, поэтому здесь он намеренно не выдуман.
9. Открыть `resources/styles/hypertension.css` и показать:
   - три точных кода цветов в `:root`;
   - `:hover` у `.stage-card` и `.next-button`;
   - `border-radius` у карточек и нижней навигации.

## 10–12. Три GET-запроса в Network

10. **Лента:** открыть `/feed/2`, затем `/feed/2?next=true`. В Network показать URL, ID `2` и параметр `next=true`.
11. **Черновик:** открыть `/draft`. Это отдельный GET, возвращающий единственную услугу со статусом `черновик`.
12. **Плитка + фильтр:** открыть `/stages?sbp=170`. Показать параметр `sbp=170` и то, что значение остается в поле input после ответа сервера.

## 13–17. Minio, Response, коллекция и шаблоны

13. На `/stages` открыть Network → document → Response. Показать `http://localhost:9000/hypertension-media/stage1.png` и `stage1.mp4` (MP4 есть в `data-video-url`).
14. Открыть `internal/app/repository/hypertension_repository.go` и показать поля `ImageURL` и `VideoURL` в коллекции.
15. Открыть `templates/hypertension_feed.html`: `poster`, `<source src=...VideoURL>`.
16. Открыть `templates/hypertension_draft.html`: `<img ...ImageURL>` и `<video ...VideoURL>`.
17. Открыть `templates/hypertension_grid.html`: `<img ...ImageURL>` и `data-video-url="...VideoURL"`.

## 18–19. Роутинг и контроллеры

18. `internal/api/hypertension_server.go` — показать ровно три `router.GET`:
   - `/feed/*serviceID`;
   - `/draft`;
   - `/stages`.
19. `internal/app/handler/hypertension_handler.go` — показать три контроллера:
   - `ShowHypertensionFeed`;
   - `ShowHypertensionDraft`;
   - `ShowHypertensionGrid`.

Для фильтрации объяснить: `ShowHypertensionGrid` читает `ctx.Query("sbp")`, преобразует его в число и вызывает `FilterPublishedHypertensionBySBP`. Фильтрация выполняется на сервере, JavaScript отсутствует.
