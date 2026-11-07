# taskService

TaskFlow состоит из frontend, auth-service, task-service, database-service и PostgreSQL. Браузер открывает только frontend; nginx проксирует `/auth` в auth-service и `/api` в task-service.

## Локальный запуск

Требуются Docker Compose, `make`, OpenSSL и `migrate`.

```bash
cp .env.example .env
make generate-auth-keys
make deploy-service
```

После запуска интерфейс доступен на `http://localhost:3000`. Миграции применяются командой `make migrate-up` (она уже входит в `make deploy-service`). Реальные `.env` и PEM-ключи исключены из Git.

Архитектура и правила расширения авторизации описаны в [docs/auth-architecture.md](docs/auth-architecture.md).
