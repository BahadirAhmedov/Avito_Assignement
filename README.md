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



## Мягкое удаление пользователей

Пользователь не удаляется из базы физически: в `users` проставляется `deleted_at`.
Помеченная запись исключается из обычной работы сервиса:

- `GET /team/get` не возвращает её в составе команды;
- `POST /users/setIsActive` и `GET /users/getReview` отвечают `404 NOT_FOUND`;
- `POST /pullRequest/create` не берёт её в ревьюеры и не даёт создать PR от её имени
  (`404 NOT_FOUND`, `author not found`);
- `POST /pullRequest/reassign` не предлагает её как кандидата на замену, но позволяет
  снять её с уже назначенного PR — иначе удалённый ревьюер остался бы заблокирован.

Сама строка остаётся в базе вместе с историей в `pr_reviewers`, поэтому ранее созданные
PR не теряют авторов и назначенных ревьюеров.

### `POST /users/delete` — пометить пользователя удалённым

Админская ручка, требует заголовок `Authorization`.

```sh
curl -X POST localhost:8080/users/delete \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u2"}'
```

Ответ `200 OK` — состояние пользователя на момент удаления:

```json
{"user":{"user_id":"u2","username":"Bob","team_name":"core","is_active":true}}
```

| Ситуация | Статус | Код |
|---|---|---|
| пользователь не найден или уже удалён | 404 | `NOT_FOUND` |
| нет или неверный `Authorization` | 401 | `UNAUTHORIZED` |
| тело без `user_id` | 400 | `INVALID_REQUEST` |
| прочее | 500 | `INTERNAL_ERROR` |

### `POST /users/restore` — восстановить удалённого пользователя

```sh
curl -X POST localhost:8080/users/restore \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"u2"}'
```

Ответ `200 OK` — восстановленный пользователь, с прежней командой и флагом `is_active`:

```json
{"user":{"user_id":"u2","username":"Bob","team_name":"core","is_active":true}}
```

| Ситуация | Статус | Код |
|---|---|---|
| пользователя нет в базе | 404 | `NOT_FOUND` |
| пользователь есть, но не был удалён | 409 | `USER_NOT_DELETED` |
| нет или неверный `Authorization` | 401 | `UNAUTHORIZED` |
| тело без `user_id` | 400 | `INVALID_REQUEST` |
| прочее | 500 | `INTERNAL_ERROR` |
