# Warehouse Control

Минимальная система управления складом: товары, история изменений, роли (user / manager / admin) и веб‑интерфейс.

## Что внутри
- Backend: Go (Gin), слои repo → usecase → delivery (HTTP)
- Frontend: статическая страница `frontend/` (Bootstrap), доступна по `/ui`
- История: фиксирует INSERT/UPDATE/DELETE с `old_data`/`new_data`
- Роли: user (чтение), manager (чтение/запись), admin (полные права)
- OpenAPI: `api/openapi.yaml`

## Быстрый старт
1) Конфигурация
- Файл: `config/app.yaml`
- Env: `JWT_SECRET` — секрет для подписи JWT, `POSTGRES_USER` — пользователь БД, `POSTGRES_PASSWORD` — пароль, `POSTGRES_DB` — имя БД

2) Docker Compose
```bash
touch .env
echo "JWT_SECRET=<secret>" >> .env
echo "POSTGRES_USER=<user>" >> .env
echo "POSTGRES_PASSWORD=<password>" >> .env
echo "POSTGRES_DB=warehouse" >> .env
docker compose up --build -d
```

3) Открыть UI и API
- UI: `http://localhost:8080/ui/`
- API: `http://localhost:8080/api/v1`
- Health: `GET /api/v1/meta/health`

## Работа с ролями в UI
- В шапке выбор роли и кнопки создания нового пользователя по роли (+ user/manager/admin).
- При выборе роли сначала используется кеш токена, иначе выполнится `POST /api/v1/meta/jwt/{role}`.

## OpenAPI
- Файл спецификации: `api/openapi.yaml`
- База: `/api/v1`
- Основные эндпоинты:
  - `POST /items` — создать товар
  - `GET /items` — список товаров
  - `PUT /items/{id}` — обновить товар
  - `DELETE /items/{id}` — удалить товар
  - `POST /meta/jwt/{role}` — выдать JWT (1=user, 2=manager, 3=admin)
  - `GET /meta/history` — история, фильтры: даты, действие, user_id, item_id, роль

## Тесты
```bash
go test ./...
```

## Ручной запуск (без Docker)
```bash
touch .env
echo "JWT_SECRET=<secret>" >> .env
echo "POSTGRES_USER=<user>" >> .env
echo "POSTGRES_PASSWORD=<password>" >> .env
echo "POSTGRES_DB=warehouse" >> .env
# Запустите Postgres и примените миграции из migrations/
go run ./src/cmd/app
```

## Структура
- `src/internal/repo` — работа с БД
- `src/internal/service` — бизнес‑логика
- `src/internal/delivery` — HTTP сервер и хэндлеры
- `frontend/` — статические файлы UI
- `migrations/` — SQL миграции

Подсказка: UI шлёт `Authorization: Bearer <jwt>`. Получить токен можно через меню ролей в интерфейсе.


