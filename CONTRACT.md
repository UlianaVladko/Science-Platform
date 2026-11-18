# Контракт BFF ↔ Frontend

BFF собирает данные из трёх источников и всегда явно сообщает, откуда пришла каждая сущность.

| Источник | Что отдаёт | Поле-маркер |
|---|---|---|
| `upstream` | 20 эндпоинтов FastAPI `/api/v1` | `source: "api"` |
| `synthetic` | расширенный демо-корпус, 20 витрин дашборда | `source: "demo"` |
| `state` | Q&A, лайки, подписки, факты дисциплин (BFF, in-memory) | `source: "bff"` |

Базовый префикс — `/api`. Все ответы JSON в UTF-8, кроме файлов публикации.
Ошибка всегда в едином конверте:

```json
{ "error": { "code": "PUBLICATION_NOT_FOUND", "message": "...", "details": null, "request_id": "req-..." } }
```

## Группа 1. Каталог и дисциплины

### `GET /api/disciplines`
Список дисциплин верхнего уровня.

```json
{
  "items": [
    {
      "id": "cat-cs", "slug": "computer-science", "name": "Информатика",
      "icon": "cpu", "description": "Вычислительные методы, алгоритмы, системы",
      "accent": "#6d5efc",
      "publications_count": 42, "subcategories_count": 3, "source": "api"
    }
  ],
  "source": "api"
}
```

- `icon` — идентификатор из закрытого набора, фронт рисует inline-SVG.
  Допустимые: `cpu`, `dna`, `gear`, `atom`, `flask`, `globe`, `chart`, `brain`, `shield`, `leaf`, `building`, `math`.
- `accent` — цвет акцента карточки, HEX. Подбирается детерминированно по `slug`.
- `publications_count` включает подкатегории.

### `GET /api/disciplines/{slug}/subcategories`
Подкатегории одной дисциплины; `publications_count` — только собственный.

### `GET /api/disciplines/{slug}/fun-fact`
Один факт, выбирается псевдослучайно из пула дисциплины (стабильно в пределах суток).

```json
{ "discipline_slug": "biology", "fact_type": "growth_yoy",
  "text": "В биологии в этом году на 23% больше публикаций, чем в прошлом",
  "value": 23, "period": "2025 → 2026", "source": "bff" }
```

`fact_type`: `growth_yoy` | `top_keyword` | `most_cited` | `new_authors` | `avg_coauthors`.
Если данных нет — `200` с `fact_type: "fallback"` и нейтральным текстом.

## Группа 2. Поиск

### `POST /api/search`
Тело — расширение контракта upstream (`POST /api/v1/search`):

```json
{
  "query": "нейросети",
  "entity_type": "publication",
  "mode": "hybrid",
  "filters": { "category": "biology", "year_from": 2015, "year_to": 2026,
               "publication_type": ["article"], "language": "ru", "author": "author-1" },
  "sort": "relevance",
  "page": 1, "page_size": 20,
  "include_facets": true,
  "scope": "all"
}
```

- `scope`: `api` | `demo` | `all` (по умолчанию `all`). `all` объединяет результаты upstream и демо-корпуса, дедуп по `id`, порядок по `score`.
- `filters.category` принимает и `cat-*`, и `slug` — BFF нормализует.
- `sort`: `relevance` | `date` | `citations` | `views` | `downloads`.

Ответ:

```json
{
  "items": [ { "entity_type": "publication", "score": 0.92, "source": "api",
               "entity": { "id": "pub-1", "...": "PublicationSummary" },
               "snippets": ["...текст вокруг совпадения..."] } ],
  "facets": [
    { "field": "language", "buckets": [{ "value": "ru", "count": 34 }] },
    { "field": "category", "buckets": [...], "labels": { "cat-nlp": "Обработка текста" } }
  ],
  "total": 154, "page": 1, "page_size": 20,
  "took_ms": 12, "degraded": false
}
```

- `degraded: true` — upstream недоступен, выдача собрана только по демо-корпусу.
- Фасеты считаются до пагинации. `labels` заполняется только для `category`.

### `GET /api/search/suggest?q=...&limit=10`
Объединённые подсказки upstream и демо-корпуса:

```json
{ "items": [ { "type": "publication", "text": "Нейронные сети...", "id": "pub-1",
               "matches_count": 3, "source": "api" } ] }
```

`type`: `publication` | `author` | `term`. `id` присутствует у публикаций и авторов для прямого перехода.

## Группа 3. Публикации

### `GET /api/publications/{id}`
Карточка. Для `pub-*` проксируется upstream (`Publication` + `source: "api"`), для демо-ID дополняется:

```json
{ "id": "demo-17", "...": "поля Publication",
  "likes_count": 12, "liked": false, "source": "demo" }
```

Ответ `404`, если `id` неизвестен обоим источникам.

### `GET /api/publications/{id}/references?direction=outgoing|incoming`
`Page<Reference>` как в upstream, с `source` у каждой ссылки.

### `GET /api/publications/{id}/related?limit=5`
Похожие публикации (по понятиям, ключевым словам и категории).

```json
{ "items": [ { "publication": { "...": "PublicationSummary" }, "score": 0.71, "shared": ["Neural networks"] } ] }
```

### `POST /api/publications/{id}/view` → `{ "publication_id": "...", "views": 11 }`
Регистрирует просмотр в BFF и пробрасывает upstream, если публикация из API.

### `POST /api/publications/{id}/like` / `DELETE /api/publications/{id}/like`
```json
{ "publication_id": "demo-17", "liked": true, "likes_count": 13 }
```
Идемпотентно: повторный like не увеличивает счётчик.

### `PUT /api/publications/{id}/files/{kind}` / `GET ...`
Проксируется в upstream без изменений. Для демо-публикаций — `404 FILE_NOT_FOUND`.

### `POST /api/publications`
Создание. BFF проксирует в upstream; при недоступности upstream создаёт запись в демо-корпусе и возвращает `201` с `source: "demo"`.

### `GET /api/publications/lookup?doi=...&title=...`
Автозаполнение формы. Источник данных — не задан внешний сервис, поэтому отдаётся детерминированное заполнение из демо-метаданных:

```json
{ "found": true, "query": "10.5555/example.100", "source": "demo-mock",
  "title": "...", "authors": [{ "name": "...", "affiliation": "..." }],
  "year": 2024, "abstract": "...", "language": "ru", "keywords": ["..."],
  "publication_type": "article", "category_slug": "machine-learning" }
```

`found: false` при неизвестном DOI (тот же конверт без полей заполнения).

## Группа 4. Авторы

### `GET /api/authors/{id}`
Проксируется из upstream (`Author`), плюс `source: "api"`.

### `GET /api/authors/{id}/stats`
```json
{ "publications_by_year": [{ "year": 2020, "count": 5 }],
  "citations_by_year": [{ "year": 2020, "count": 12 }],
  "top_categories": [{ "slug": "biology", "name": "Биология", "count": 20 }],
  "coauthors": [{ "id": "author-2", "name": "Петров П.", "count": 8 }] }
```

### `GET /api/authors/{id}/publications?page=1&page_size=20`
`Page<PublicationSummary>`.

### `GET /api/authors/{id}/coauthor-graph?depth=2`
```json
{ "nodes": [{ "id": "author-1", "label": "Анна Петрова", "type": "author",
              "weight": 45, "category": "cat-cs" }],
  "edges": [{ "source": "author-1", "target": "author-2", "weight": 8 }],
  "truncated": false }
```

## Группа 5. Аналитика и графы

### `GET /api/analytics/summary`
KPI корпуса:
```json
{ "total_publications": 214, "total_authors": 61, "total_citations": 1832,
  "total_downloads": 4401, "avg_citations_per_publication": 8.56,
  "publications_this_period": 38, "new_authors_this_period": 9,
  "period": "2026", "data_as_of": "2026-11-18T00:00:00Z" }
```

### `GET /api/analytics/top?entity=publication|author|keyword|concept&metric=citations|views|downloads|publications|frequency&category=...&period=...&limit=10`
Рейтинг (формат upstream `Ranking`), считается по объединённому корпусу.

### `GET /api/dashboard/{widget}`
Одна ручка на виджет, все принимают `category` (slug или `cat-*`) и `period` (`all`|`2024`|`2025`|`2026`).
Общий конверт ответа:

```json
{ "widget": "publications-timeline", "source": "demo",
  "data_as_of": "2026-11-18T00:00:00Z", "ready": true, "payload": { "...": "..." } }
```

| `widget` | `ready` | `payload` |
|---|---|---|
| `publications-timeline` | всегда | `{ "series": [{ "year": 2015, "count": 12 }] }` |
| `intake-timeline` | всегда | `{ "series": [{ "date": "2026-01", "count": 45 }] }` |
| `by-category` | всегда | `{ "slices": [{ "slug": "biology", "name": "Биология", "count": 32, "color": "#..." }] }` |
| `top-authors` | всегда | `{ "items": [{ "id": "author-1", "name": "...", "count": 18, "citations": 240 }] }` |
| `keywords-cloud` | зависит от NLP | `{ "words": [{ "text": "crispr", "weight": 42 }] }` |
| `top-concepts` | зависит от NLP | `{ "items": [{ "id": "concept-1", "name": "...", "count": 14 }] }` |
| `lemmas-frequency` | зависит от NLP | `{ "items": [{ "lemma": "анализ", "count": 96 }] }` |
| `by-language` | всегда | `{ "slices": [{ "value": "ru", "name": "Русский", "count": 120 }] }` |
| `by-type` | всегда | `{ "slices": [{ "value": "article", "name": "Статья", "count": 90 }] }` |
| `top-viewed` | всегда | `{ "items": [...] }` |
| `top-cited` | всегда | `{ "items": [...] }` |
| `top-downloaded` | всегда | `{ "items": [...] }` |
| `by-status` | всегда | `{ "slices": [{ "value": "published", "count": 190 }] }` |
| `by-source` | всегда | `{ "slices": [{ "value": "crossref", "count": 120 }] }` |
| `coauthors-distribution` | всегда | `{ "buckets": [{ "bucket": "2", "count": 40 }] }` |
| `citation-graph` | всегда | `{ "nodes": [...], "edges": [...], "truncated": false }` |
| `summary` | всегда | то же, что `/api/analytics/summary` |
| `by-affiliation` | всегда | `{ "items": [{ "name": "МГУ", "count": 34 }] }` |
| `activity-heatmap` | всегда | `{ "days": ["2026-01-01"], "categories": ["Биология"], "values": [[0,0,3]] }` |
| `citations-timeline` | всегда | `{ "series": [{ "year": 2015, "count": 210 }] }` |
| `top-publications` | всегда | `{ "metric": "citations", "items": [...] }` |

`ready: false` означает «Данные готовятся» — фронт показывает состояние-заглушку с пояснением.
Флаг управляется переменной окружения BFF `SYNTHETIC_NLP_READY` (по умолчанию `false`, как в UI-документе: widgets 5–7 требуют NLP-пайплайн).

### `GET /api/graphs?type=citation|coauthor|concept&root_id=...&max_depth=2&max_nodes=50&category=...&period=...`
Проксируется из upstream при доступности. Если `root_id` указывает на демо-сущность — граф строится по демо-корпусу. Формат — `Graph` upstream плюс `source`.

## Группа 6. Вопросы и ответы (BFF-состояние)

| Метод | Путь | Ответ |
|---|---|---|
| `GET` | `/api/questions?page=1&page_size=10&category=...&sort=newest\|top` | `Page<QuestionSummary>` |
| `POST` | `/api/questions` | `201 Question`, тело `{ "title", "body", "category_slug" }` |
| `GET` | `/api/questions/{id}` | `Question` с `answers` |
| `POST` | `/api/questions/{id}/answers` | `201 Answer`, тело `{ "body", "author_name" }` |
| `POST` | `/api/answers/{id}/vote` | `{ "id", "votes" }`, тело `{ "vote": 1 \| -1 }` |

```json
{
  "id": 1, "title": "Как работает CRISPR-Cas9?",
  "author": { "id": "author-1", "name": "Анна Петрова" },
  "category": { "slug": "biology", "name": "Биология" },
  "answers_count": 5, "views_count": 320, "votes": 4,
  "created_at": "2026-09-01T00:00:00Z", "source": "bff",
  "answers": [ { "id": 1, "body": "...", "author": { "id": "author-2", "name": "..." },
                 "votes": 12, "created_at": "2026-09-02T00:00:00Z" } ]
}
```

Валидация: `title` 5–200 символов, `body` 1–5000, `author_name` 1–120, `vote` ∈ {1, −1}.
Голос одного `voter` (передаётся в `X-Voter`) учитывается один раз: повторный голос заменяет предыдущий.

## Группа 7. Подписки (BFF-состояние)

| Метод | Путь | Ответ |
|---|---|---|
| `GET` | `/api/users/{id}/profile` | профиль автора для карточки |
| `POST` | `/api/users/{id}/follow` | `{ "following": true, "followers_count": 231 }` |
| `DELETE` | `/api/users/{id}/follow` | `{ "following": false, "followers_count": 230 }` |

`X-Viewer` (строка) идентифицирует пользователя для лайков и подписок; по умолчанию `guest`.

## Группа 8. Служебные

| Путь | Ответ |
|---|---|
| `GET /api/health` | `{ "status": "ok", "upstream": "up" \| "down", "corpus": 214, "version": "1.0.0" }` |
| `GET /api/meta` | `{ "title", "version", "sources": {...}, "widgets_ready": {...} }` |
