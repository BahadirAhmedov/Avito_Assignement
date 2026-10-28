## Запуск проекта через `docker-compose up`

### Шаг 1 нужно склонировать репозиторий 

- git clone https://github.com/BahadirAhmedov/Avito_Assignement

### Шаг 2 запустить сервис и зависимости

- docker-compose up --build

### Шаг 3 после запуска Api будет достпуна по адресу
-  http://localhost:8080




## Запуск проекта через `Makefile`
local.env файл для запуска через Makefile

ENV=local\
ADDRES=:8080\
DB_HOST=localhost\
DB_PORT=5433\
DB_USER=postgres\
DB_PASSWORD=postgres\
DB_NAME=avito



### Шаг 1 нужно собрать проект

- make build-service

### Шаг 2 запуск
- make run


### Применение миграция 
- make run-migrator


## Мягкое удаление пользователей (soft delete)

Пользователи не удаляются физически: в таблице `users` проставляется `deleted_at`.
Запись остаётся в базе, но исключается из обычных выборок:

- `GET /team/get` — удалённые не попадают в `members`
- `GET /users/getReview` — для удалённого отдаёт `404 NOT_FOUND`
- `POST /users/setIsActive` — для удалённого отдаёт `404 NOT_FOUND`
- `POST /pullRequest/create` — удалённые не назначаются ревьюерами, а автор-удалённый даёт `404 NOT_FOUND`
- `POST /pullRequest/reassign` — удалённые не предлагаются как замена

Открытые PR, где удалённый уже был назначен ревьюером, не меняются автоматически —
его можно заменить через `POST /pullRequest/reassign`.

### `POST /users/delete`

Требует заголовок `Authorization: Bearer secret`.

```bash
curl -X POST localhost:8080/users/delete \
  -H "Authorization: Bearer secret" \
  -d '{"user_id":"u2"}'
```

```json
{"user":{"user_id":"u2","username":"Bob","team_name":"core","is_active":true,"deleted_at":"2026-10-28T20:41:05.777149Z"}}
```

Коды ошибок: `401 UNAUTHORIZED`, `400 INVALID_REQUEST`, `404 NOT_FOUND`, `409 USER_DELETED` (уже удалён).

### `POST /users/restore`

Требует заголовок `Authorization: Bearer secret`.

```bash
curl -X POST localhost:8080/users/restore \
  -H "Authorization: Bearer secret" \
  -d '{"user_id":"u2"}'
```

```json
{"user":{"user_id":"u2","username":"Bob","team_name":"core","is_active":true}}
```

Коды ошибок: `401 UNAUTHORIZED`, `400 INVALID_REQUEST`, `404 NOT_FOUND`, `409 USER_NOT_DELETED` (пользователь не был удалён).
