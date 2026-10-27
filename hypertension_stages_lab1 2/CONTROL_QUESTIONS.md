# Контрольные вопросы — короткий ответ для защиты

## MVT

MVT = Model–View–Template, архитектурный подход Django-подобного типа. **Model** хранит и получает данные, **View** содержит обработчик/бизнес-логику запроса, **Template** отвечает за HTML-представление. В этой лабораторной аналог Model — `HypertensionRepository` и коллекция `hypertensionServices`, View — методы `ShowHypertension...`, Template — три HTML-файла.

## MVC в этой лабораторной

- **Model:** `HypertensionStage`, `HypertensionRepository`, коллекция услуг.
- **Controller:** `HypertensionHandler`; принимает HTTP-запрос, читает `id`, `next`, `sbp`, фильтрует данные и считает лайки.
- **View:** Go HTML templates `hypertension_feed.html`, `hypertension_draft.html`, `hypertension_grid.html` + CSS.

В терминологии Gin здесь удобнее говорить именно MVC: handler выступает контроллером, а HTML template — представлением.

## Шаблонизация

Шаблонизация — генерация HTML из шаблона и данных. Gin использует Go `html/template`. Например `{{ .Card.Service.SystolicBP }}` подставляет значение САД, а `{{ range .Cards }}` создает карточку для каждого элемента списка. `html/template` экранирует текстовые данные, что снижает риск XSS.

## HTTP и модель OSI

HTTP — протокол прикладного уровня. В OSI это **7-й уровень (Application)**. HTTP обычно работает поверх TCP (транспортный уровень; в HTTP/3 — QUIC поверх UDP), ниже находятся IP, канальный и физический уровни. В лабораторной браузер отправляет GET-запрос, сервер Gin возвращает HTTP-response с HTML.

Основные части HTTP-запроса: метод (`GET`), URL, headers, необязательное body. У GET параметры поиска обычно находятся в query string, например `/stages?sbp=170`.

## Web

Web — система ресурсов, связанных URL/URI и доступных через сетевые протоколы, прежде всего HTTP/HTTPS. Браузер — клиент, Gin-приложение — web-сервер приложения, Minio — отдельный HTTP-сервис для объектов изображений и видео.

## HTML

HTML — язык разметки веб-документов. Он задает семантическую структуру страницы: `main`, `section`, `nav`, `form`, `input`, `video`, `img`, ссылки и т. д. CSS отвечает за внешний вид. В этой лабораторной JavaScript не используется: переходы выполняются ссылками, поиск — обычной GET-формой, видео запускается HTML-атрибутом `autoplay`.
