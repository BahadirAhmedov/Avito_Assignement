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
Удалённые пользователи остаются в базе, но исключаются из обычной работы сервиса:

- не возвращаются в `GET /team/get`
- не назначаются ревьюерами в `POST /pullRequest/create` и `POST /pullRequest/reassign`
- не могут быть автором PR (`AUTHOR_NOT_FOUND`)
- для них `POST /users/setIsActive` и `GET /users/getReview` отвечают `404 NOT_FOUND`

Уже существующие назначения в `pr_reviewers` сохраняются — заменить такого ревьюера
можно через `POST /pullRequest/reassign`.

### `POST /users/delete`

Помечает пользователя удалённым. Требует заголовок `Authorization: Bearer secret`.

```
curl -X POST localhost:8080/users/delete \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u3"}'
```

Ответ `200`:

```json
{"user":{"user_id":"u3","username":"Carol","team_name":"core","is_active":true,"deleted_at":"2025-10-28T20:51:21.675641Z"}}
```

Ошибки: `401 UNAUTHORIZED`, `400 INVALID_REQUEST`, `404 NOT_FOUND` (пользователя нет или он уже удалён).

### `POST /users/restore`

Восстанавливает ранее удалённого пользователя (`deleted_at = NULL`). Требует заголовок `Authorization: Bearer secret`.

```
curl -X POST localhost:8080/users/restore \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u3"}'
```

Ответ `200`:

```json
{"user":{"user_id":"u3","username":"Carol","team_name":"core","is_active":true}}
```

Ошибки: `401 UNAUTHORIZED`, `400 INVALID_REQUEST`, `404 NOT_FOUND` (пользователя нет),
`400 USER_NOT_DELETED` (пользователь не был удалён).

Повторное добавление удалённого пользователя через `POST /team/add` также снимает пометку удаления.
