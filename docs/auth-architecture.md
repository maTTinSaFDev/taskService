# Архитектура авторизации TaskFlow

## Компоненты и границы ответственности

```text
Browser
  │ same-origin /auth, /api
  ▼
frontend nginx ── /auth ──► auth-service ── gRPC ──► database-service ──► PostgreSQL
       │                         │ private Ed25519 key
       └──── /api ───────► task-service
                                  │ public Ed25519 key
                                  └── gRPC ─────────► database-service
```

- **auth-service** валидирует email и пароль, выполняет bcrypt, выпускает access JWT и управляет lifecycle refresh-сессий.
- **database-service** — единственный владелец PostgreSQL. Внутренний `AuthPersistenceService` хранит credential и refresh-сессии, но не проверяет пароли и не выпускает токены.
- **task-service** не вызывает auth-service. Он проверяет подпись, issuer, audience и срок access JWT локально, затем использует `sub` как `user_id`.
- **frontend** хранит access-токен только в памяти. Refresh-токен недоступен JavaScript и передаётся cookie `HttpOnly; SameSite=Lax; Path=/auth`.

## Последовательности

### Register

1. Browser отправляет `POST /auth/register` с `name`, `email`, `password`.
2. auth-service нормализует email, проверяет пароль (12–72 байта) и создаёт bcrypt-хеш.
3. database-service атомарно создаёт пользователя; уникальный индекс по `LOWER(email)` исключает дубликаты.
4. auth-service создаёт случайный 256-битный refresh-токен. В БД передаётся только SHA-256 hash.
5. Browser получает access JWT в JSON и refresh-cookie.

### Login

1. `POST /auth/login` получает credential через внутренний gRPC API.
2. bcrypt сравнивается в auth-service; клиенту возвращается единая ошибка для неизвестного email и неверного пароля.
3. Создаются новая refresh-сессия и access JWT.

### Refresh

1. Browser отправляет `POST /auth/refresh`; cookie прикладывается автоматически.
2. auth-service хеширует cookie и генерирует replacement token.
3. database-service в транзакции блокирует активную сессию, отзывает её и создаёт replacement. Повторное использование старого токена отклоняется.
4. Browser получает новый access JWT и заменённую refresh-cookie.

### Protected request

1. Frontend добавляет `Authorization: Bearer <access>` к `/api`.
2. task-service локально проверяет Ed25519-подпись, `iss`, `aud`, `iat`, `nbf`, `exp` и положительный `sub`.
3. `user_id` из body/query не используется для авторизации. Создание привязывается к `sub`; чтение, изменение и удаление проверяют ownership.
4. При `401` frontend один раз выполняет refresh и повторяет исходный запрос.

### Logout

1. `POST /auth/logout` отзывает текущий hash refresh-сессии.
2. Сервер очищает cookie; frontend удаляет access-токен из памяти.

## Ключи и переменные окружения

Используется Ed25519. Приватный ключ монтируется только в auth-service, публичный — только в task-service. Сгенерировать пару:

```bash
make generate-auth-keys
```

Основные переменные: `JWT_ISSUER`, `JWT_AUDIENCE`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL`, `AUTH_COOKIE_SECURE`, `JWT_PRIVATE_KEY_FILE`, `JWT_PUBLIC_KEY_FILE`. Для HTTPS необходимо `AUTH_COOKIE_SECURE=true`.

## Правила доступа

- Доски принадлежат `owner_user_id == JWT sub`.
- Задачи в текущей модели считаются личными: `task.user_id == JWT sub`.
- Колонка может изменяться только через принадлежащую пользователю доску; database-service дополнительно проверяет `board_id` при удалении.
- Создание/перемещение задачи разрешено только в колонку доски того же пользователя.
- Список пользователей и чтение отображаемого имени сохранены для интерфейса. Изменение и удаление пользователя разрешены только самому пользователю; создание пользователя выполняется через регистрацию.

## Подключение нового сервиса

1. Передать контейнеру только `auth-public.pem`, `JWT_ISSUER` и `JWT_AUDIENCE`.
2. Проверять JWT локально алгоритмом EdDSA и разрешать только ожидаемые issuer/audience.
3. Помещать положительный `sub` в request context.
4. Не принимать user identity из клиентского payload и проверять ownership на границе use case/persistence.

## Известные компромиссы первой версии

- Межсервисный gRPC внутри compose пока без TLS/mTLS. Для запуска в недоверенной сети нужен service mesh или mTLS.
- Нет ролей, MFA, rate limiting и password reset. Login/register следует ограничивать на ingress перед публичным развёртыванием.
- При компрометации refresh-токена ротация блокирует повторное использование старой сессии, но пока не отзывает всю «семью» токенов. Следующий шаг — `family_id` и revoke-on-reuse.
- Старые пользователи из миграции остаются без email/password и не могут войти, пока не появится отдельный flow привязки credential.
