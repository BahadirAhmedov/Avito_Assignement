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

Пользователь не удаляется из БД физически: в `users` есть колонка `deleted_at`
(миграция `3_init`). Запись с непустым `deleted_at` считается удалённой и исключается из
обычной работы сервиса:

- `GET /team/get` не показывает её в составе команды;
- `GET /users/getReview` отвечает `404 NOT_FOUND`;
- `POST /users/setIsActive` отвечает `404 NOT_FOUND`;
- удалённый пользователь не попадает в автор-/ревьюер-выборки `POST /pullRequest/create`
  и в кандидаты замены `POST /pullRequest/reassign` (заменить самого удалённого ревьюера
  на живого по-прежнему можно).

Вернуть пользователя можно только через `POST /users/restore` — повторный
`POST /team/add` с тем же `user_id` удалённую запись не воскрешает.

Обе ручки административные: требуется заголовок `Authorization: Bearer secret`.

### `POST /users/delete`

```sh
curl -X POST localhost:8080/users/delete \
  -H "Authorization: Bearer secret" -H "Content-Type: application/json" \
  -d '{"user_id":"u1"}'
```

```json
{
  "user": {
    "user_id": "u1",
    "username": "Ivan",
    "team_name": "backend",
    "is_active": true
  }
}
```

| Код ответа | `code` | Когда |
|---|---|---|
| 200 | — | пользователь помечен удалённым |
| 400 | `INVALID_REQUEST` | тело не разобрано, нет `user_id` |
| 401 | `UNAUTHORIZED` | нет или неверный `Authorization` |
| 404 | `NOT_FOUND` | пользователя нет в БД |
| 409 | `USER_ALREADY_DELETED` | пользователь уже удалён |
| 500 | `INTERNAL_ERROR` | непредвиденная ошибка |

### `POST /users/restore`

```sh
curl -X POST localhost:8080/users/restore \
  -H "Authorization: Bearer secret" -H "Content-Type: application/json" \
  -d '{"user_id":"u1"}'
```

Тело ответа — та же структура `{"user": {...}}`.

| Код ответа | `code` | Когда |
|---|---|---|
| 200 | — | пользователь восстановлен |
| 400 | `INVALID_REQUEST` | тело не разобрано, нет `user_id` |
| 401 | `UNAUTHORIZED` | нет или неверный `Authorization` |
| 404 | `NOT_FOUND` | пользователя нет в БД |
| 409 | `USER_NOT_DELETED` | пользователь не был удалён |
| 500 | `INTERNAL_ERROR` | непредвиденная ошибка |


## Выгрузка своих данных (`GET /users/export`)

Ручка собирает всё, что сервис знает о пользователе, в один файл: профиль из `users`,
живых коллег по команде из `users`, авторские пулл-реквесты из `pull_requests` и назначения
на ревью из `pr_reviewers`.

В отличие от `POST /users/delete` и `POST /users/restore`, это **не** административная ручка:
в `Authorization: Bearer <user_id>` передаётся идентификатор самого пользователя, и выгрузить
можно только собственные данные — при `user_id` в запросе, не совпадающем с токеном, ответ
`403 FORBIDDEN`. Админский `Bearer secret` здесь доступа не даёт. Удалённый (soft delete)
пользователь не выгружается — `404 NOT_FOUND`.

Формат задаётся `format=json` (по умолчанию) или `format=csv`; в обоих случаях ответ приходит
с `Content-Disposition: attachment`, то есть сохраняется как файл `user_<user_id>_export.<json|csv>`.

```sh
curl -OJ "localhost:8080/users/export?user_id=u1" -H "Authorization: Bearer u1"
curl -OJ "localhost:8080/users/export?user_id=u1&format=csv" -H "Authorization: Bearer u1"
```

```json
{
  "user": {
    "user_id": "u1",
    "username": "Ivan",
    "team_name": "backend",
    "is_active": true
  },
  "teammates": [
    {
      "user_id": "u2",
      "username": "Olga",
      "team_name": "backend",
      "is_active": true
    }
  ],
  "authored_pull_requests": [
    {
      "pull_request_id": "pr1",
      "pull_request_name": "Add export",
      "author_id": "u1",
      "status": "OPEN"
    }
  ],
  "review_pull_requests": [
    {
      "pull_request_id": "pr2",
      "pull_request_name": "Fix logs",
      "author_id": "u2",
      "status": "OPEN"
    }
  ]
}
```

CSV — один плоский файл на все разделы, колонка `section` говорит, откуда строка:

```csv
section,user_id,username,team_name,is_active,pull_request_id,pull_request_name,author_id,status
profile,u1,Ivan,backend,true,,,,
teammate,u2,Olga,backend,true,,,,
authored_pull_request,,,,,pr1,Add export,u1,OPEN
review_pull_request,,,,,pr2,Fix logs,u2,OPEN
```

| Код ответа | `code` | Когда |
|---|---|---|
| 200 | — | файл выгрузки отдан |
| 400 | `INVALID_REQUEST` | нет `user_id`, либо `format` не `json`/`csv` |
| 401 | `UNAUTHORIZED` | нет заголовка `Authorization` или он не вида `Bearer <user_id>` |
| 403 | `FORBIDDEN` | запрошены чужие данные: `user_id` не совпадает с токеном |
| 404 | `NOT_FOUND` | пользователя нет в БД или он удалён |
| 500 | `INTERNAL_ERROR` | непредвиденная ошибка |
