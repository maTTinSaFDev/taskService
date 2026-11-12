# TaskFlow

TaskFlow — учебный аналог Microsoft To Do с канбан-досками, задачами и отдельным сервисом авторизации. Проект построен как несколько сервисов: публичный HTTP API отделён от сервиса, который единолично работает с PostgreSQL.

## Возможности

- регистрация, вход, обновление и отзыв сессии;
- несколько досок на пользователя;
- создание, переименование, перестановка и удаление колонок;
- перенос задач между колонками, прогресс, приоритет, дата и длительность;
- защита данных по владельцу из JWT, без доверия к `user_id` клиента;
- единый запуск всех компонентов через Docker Compose.

## Архитектура

```text
Browser
  │
  ▼
frontend (nginx :3000)
  ├── /auth ──► auth-service (:8090) ── gRPC ──┐
  │                   │                         │
  │                   └── Redis                ▼
  └── /api ───► task-service (:8080) ── gRPC ─► database-service (:50051)
                                                    │
                                                    ▼
                                                PostgreSQL
```

| Компонент | Ответственность |
| --- | --- |
| `frontend` | SPA и same-origin reverse proxy для `/auth` и `/api` |
| `auth-service` | bcrypt, access JWT, refresh-сессии и HTTP API авторизации |
| `task-service` | публичный HTTP API, проверка JWT и ownership |
| `database-service` | внутренний gRPC API и единственная точка доступа к PostgreSQL |
| PostgreSQL | пользователи, credentials, доски, колонки и задачи |
| Redis | SHA-256-хеши refresh-токенов с TTL |

Сервисы написаны на Go, взаимодействие `task-service`/`auth-service` с `database-service` идёт по gRPC. HTTP-маршрутизация реализована через Gorilla Mux, PostgreSQL-доступ — через pgx. Подробнее поток авторизации описан в [docs/auth-architecture.md](docs/auth-architecture.md).

## Быстрый запуск

Понадобятся Docker с Compose, `make` и утилита [golang-migrate](https://github.com/golang-migrate/migrate). Локальная установка Go для контейнерного запуска не обязательна.

```bash
cp .env.example .env
make generate-auth-keys
make deploy-service
```

После запуска:

- интерфейс: `http://localhost:3000`;
- task-service напрямую: `http://localhost:8080`;
- auth-service напрямую: `http://localhost:8090`;
- внутренний gRPC database-service: `localhost:50051`.

`make deploy-service` собирает контейнеры, ждёт прохождения healthcheck и применяет SQL-миграции. Реальные `.env` и PEM-ключи исключены из Git.

### Основные команды

| Команда | Назначение |
| --- | --- |
| `make deploy-service` | собрать и запустить весь Compose-проект, применить миграции |
| `make migrate-up` | применить неприменённые SQL-миграции |
| `make generate-auth-keys` | создать Ed25519 private/public key pair в `secrets/` |
| `make generate-grpc-code-database-service` | перегенерировать Go-код из `.proto` |
| `make test` | запустить тесты во всех четырёх Go-модулях локальным Go |
| `make test-docker` | запустить те же тесты в `golang:1.27-bookworm` |
| `docker compose down` | остановить контейнеры, сохранив именованные volumes |
| `docker compose down -v` | остановить проект и удалить локальные данные PostgreSQL/Redis |

## HTTP API

Через frontend публичные маршруты имеют базовый адрес `http://localhost:3000`. Маршруты `/api/*` требуют заголовок `Authorization: Bearer <access_token>`. Refresh-токен возвращается только в cookie `HttpOnly` и не попадает в JavaScript.

### Авторизация

| Метод и маршрут | Тело/параметры | Результат |
| --- | --- | --- |
| `POST /auth/register` | `name`, `email`, `password` | `201`, access-токен, пользователь и refresh-cookie |
| `POST /auth/login` | `email`, `password` | `200`, новая пара токенов |
| `POST /auth/refresh` | refresh-cookie | `200`, ротация refresh-сессии и новый access-токен |
| `POST /auth/logout` | refresh-cookie | `204`, отзыв текущей refresh-сессии |
| `GET /auth/me` | Bearer access-токен | текущий пользователь |

### Доски и колонки

| Метод и маршрут | Тело/параметры | Результат |
| --- | --- | --- |
| `GET /api/boards` | — | доски текущего пользователя |
| `POST /api/boards` | `title` | новая доска с начальными колонками |
| `GET /api/boards/{id}` | path `id` | доска вместе с колонками |
| `POST /api/boards/{board_id}/columns` | `title`, `position` | новая колонка |
| `PUT /api/boards/{board_id}/columns/{column_id}` | `title`, `position` | изменённая колонка |
| `PUT /api/boards/{board_id}/columns/order` | `column_ids` — полный упорядоченный список | `204`, атомарная перестановка |
| `DELETE /api/boards/{board_id}/columns/{column_id}` | необязательный `target_column_id` в query | `204`; задачи непустой колонки предварительно переносятся |

### Задачи

| Метод и маршрут | Тело/параметры | Результат |
| --- | --- | --- |
| `GET /api/tasks?board_id={id}` | необязательный фильтр доски | задачи текущего пользователя |
| `GET /api/tasks/{id}` | path `id` | одна задача |
| `POST /api/tasks` | поля задачи без `id` и `user_id` | `201`, созданная задача |
| `PUT /api/tasks` | полный объект задачи с `id` | `200`, пустое тело |
| `DELETE /api/tasks/{id}` | path `id` | `204` |

Поля задачи: `title`, `description`, `column_id`, `position`, `progress` (0–100), `is_done`, `start_time` (RFC 3339), `duration` (наносекунды Go `time.Duration`) и `priority` (1–100). `user_id` определяется только из JWT.

### Пользователь

| Метод и маршрут | Тело | Ограничение |
| --- | --- | --- |
| `GET /api/users/{id}` | — | требуется валидный access-токен |
| `PUT /api/users/{id}` | `name` | изменить можно только себя |
| `DELETE /api/users/{id}` | — | удалить можно только себя; связанные данные дают `409` |

## Примеры запросов

Ниже используется `jq` только для извлечения access-токена из JSON.

```bash
# Регистрация. Cookie сохраняется отдельно от access-токена.
ACCESS_TOKEN=$(curl -sS -c cookies.txt \
  -H 'Content-Type: application/json' \
  -d '{"name":"Vadim","email":"vadim@example.com","password":"very-secure-password"}' \
  http://localhost:3000/auth/register | jq -r '.access_token')

# Создание доски.
BOARD=$(curl -sS \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Учебный проект"}' \
  http://localhost:3000/api/boards)

echo "$BOARD" | jq
BOARD_ID=$(echo "$BOARD" | jq -r '.id')
COLUMN_ID=$(echo "$BOARD" | jq -r '.columns[0].id')

# Создание задачи.
curl -sS \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"title\":\"Дополнить README\",\"description\":\"Описать API\",\"column_id\":$COLUMN_ID,\"position\":0,\"progress\":0,\"is_done\":false,\"start_time\":\"2026-11-12T12:00:00Z\",\"duration\":3600000000000,\"priority\":20}" \
  http://localhost:3000/api/tasks | jq

# Список задач выбранной доски.
curl -sS \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "http://localhost:3000/api/tasks?board_id=$BOARD_ID" | jq

# Атомарная перестановка колонок: нужно передать все ID доски ровно один раз.
curl -i -X PUT \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"column_ids":[4,2,3,1]}' \
  "http://localhost:3000/api/boards/$BOARD_ID/columns/order"

# Обновление access-токена по HttpOnly cookie.
ACCESS_TOKEN=$(curl -sS -b cookies.txt -c cookies.txt \
  -X POST http://localhost:3000/auth/refresh | jq -r '.access_token')

# Завершение сессии.
curl -i -b cookies.txt -X POST http://localhost:3000/auth/logout
```

## Тестирование

Все unit-тесты:

```bash
make test
```

Если Go не установлен локально:

```bash
make test-docker
```

Быстрая проверка развёрнутого окружения:

```bash
docker compose ps
curl -i http://localhost:8090/healthz
curl -I http://localhost:3000
```

## Ключевые архитектурные решения

- **PostgreSQL изолирован за database-service.** Публичные сервисы не выполняют SQL и используют типизированные gRPC-контракты.
- **Access и refresh разделены.** Короткоживущий Ed25519 JWT проверяется локально в task-service; refresh-токен хранится в cookie `HttpOnly`, а в Redis сохраняется только его SHA-256-хеш с TTL.
- **Refresh ротация атомарна.** Lua-скрипт Redis заменяет сессию одной операцией, поэтому повторное использование старого токена отклоняется.
- **Идентификатор пользователя не принимается на доверии.** `user_id` берётся из JWT `sub`; сервисы дополнительно проверяют принадлежность досок, колонок и задач.
- **Порядок колонок обновляется транзакционно.** database-service блокирует доску и её колонки, проверяет полный набор ID и применяет новый порядок одной транзакцией; некорректный набор возвращает `409 Conflict`.
- **Зависимости направлены через интерфейсы.** Transport зависит от service-контрактов, service — от repository-контрактов; реализации создаются только в `main.go`. Это упрощает unit-тесты и замену транспорта или хранилища.
- **Kafka сознательно не добавлена.** Для текущего учебного объёма она не участвует в пользовательском сценарии и не усложняет локальный запуск.

## Структура репозитория

```text
frontend/                         SPA и nginx
internal/core/                    общие модели и DTO
internal/auth-service/            HTTP-авторизация, JWT, Redis sessions
internal/task-service/            публичный HTTP API и gRPC-клиенты
internal/database-service/        PostgreSQL, миграции и gRPC server
docs/auth-architecture.md         подробности security flow
docker-compose.yaml               локальная инфраструктура
Makefile                          сборка, тесты, генерация и запуск
```

## Ограничения текущей версии

- внутренний gRPC в Docker-сети пока работает без TLS/mTLS;
- нет ролей, MFA, password reset и rate limiting;
- Redis хранит снимок профиля на момент входа, поэтому после переименования он обновится при следующем login;
- удаление доски во внешнем HTTP API пока не реализовано.
