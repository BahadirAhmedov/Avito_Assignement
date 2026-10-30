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
