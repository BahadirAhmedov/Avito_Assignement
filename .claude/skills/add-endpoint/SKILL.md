---
name: add-endpoint
description: Добавление или изменение HTTP-ручки в этом сервисе — проводка через все слои (request/response DTO → sentinel-ошибка → метод storage → провайдер-интерфейс → handler → роут), включая админскую авторизацию и миграцию схемы. Использовать при задачах вида «добавь эндпоинт», «новая ручка», «новый API-метод», «добавь поле в запрос/ответ», «сделай удаление/восстановление/список для сущности X».
---

# Добавление HTTP-ручки

Инварианты проекта — в `AGENTS.md` в корне; здесь только порядок действий и то, что
специфично для ручек. Слои трогаются строго снизу вверх, иначе код не соберётся на
полпути и придётся возвращаться.

## Именование — выводится из пути роута

Путь роута и имя метода handler'а связаны однозначно, это соблюдается для всех семи
существующих ручек:

| Роут | Метод handler | Метод storage |
|---|---|---|
| `POST /team/add` | `TeamAdd` | `SaveTeam` |
| `GET /team/get` | `TeamGet` | `ReadTeam` |
| `POST /users/setIsActive` | `UsersSetIsActive` | `SetIsActive` |
| `GET /users/getReview` | `UsersGetReview` | `ReadReviewUsers` |
| `POST /pullRequest/create` | `PullRequestCreate` | `SavePullRequest` |

Правило: `<Сущность><Действие>` в handler, глагол `Save*` / `Read*` / `Set*` в storage.
Новая ручка `POST /users/delete` → handler `UsersDelete`, storage `DeleteUser`.

**Не делать:** не заводить имена вида `HandleDeleteUser`, `GetUserHandler`, `DoDelete`.

## Порядок шагов

### 1. Миграция — если нужна новая колонка

Пара `migrations/N_init.up.sql` + `migrations/N_init.down.sql`, `N` = следующий номер.
Существующие файлы не редактировать. Проверка: `make run-migrator` проходит, затем
`psql ... -c "\d users"` показывает ожидаемую схему.

### 2. DTO запроса и ответа

Запрос — в `internal/transport/http/requests/<сущность>-request.go`, ответ — в
`internal/transport/http/responses/<сущность>-response.go`. Структура запроса —
`<Сущность><Действие>Request`, все обязательные поля с `binding:"required"`.

Флаг в запросе объявляется как `*bool` (AGENTS.md §7) — иначе `false` будет отвергнут с 400.

Если ответ повторяет уже существующий (например, отдаёт одного пользователя) —
переиспользовать `responses.UserResponse`, а не плодить новый тип.

### 3. Sentinel-ошибка — если появляется новая бизнес-ситуация

Добавить в `internal/storage/storage.go` рядом с остальными:

```go
ErrUserNotDeleted = errors.New("user is not deleted")
```

Сразу же добавить парную константу кода в блок в
`internal/http-server/handlers/pullRequestService.go`:

```go
USER_NOT_DELETED = "USER_NOT_DELETED"
```

**Не делать:** не возвращать из storage ошибку, под которую нет sentinel'а — handler не
сможет её отличить от 500 и ответит пятисоткой на штатную ситуацию.

### 4. Метод storage

В `internal/storage/postgres/<сущность>.go`. Первая строка — `const op`. Ошибки драйвера
переводятся в sentinel'ы, наружу уходит только `fmt.Errorf("%s: %w", op, ...)`.

Одиночный statement — через `s.db`, несколько записей — транзакция (AGENTS.md §5).
Для «обновить и вернуть» использовать `UPDATE ... RETURNING` в один запрос:

```go
err := s.db.QueryRow(
    `UPDATE users SET deleted_at = NOW(), updated_at = NOW()
         WHERE user_id = $1 AND deleted_at IS NULL
         RETURNING user_id, username, team_name, is_active`,
    request.UserId,
).Scan(&user.UserId, &user.UserName, &user.TeamName, &user.IsActive)

if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
    }
    return models.User{}, fmt.Errorf("%s: %w", op, err)
}
```

Если метод возвращает срез — нормализовать `nil` в пустой срез перед `return` (AGENTS.md §6).

**Внимание на `team_name`:** колонка nullable (`ON DELETE SET NULL`), а существующий код
сканирует её в `string`. Если новый запрос может вернуть пользователя без команды —
сканировать в `sql.NullString`, иначе `Scan` упадёт.

### 5. Провайдер-интерфейс

Три интерфейса (`TeamProvider`, `UserProvider`, `PullRequestProvider`) объявлены в
`internal/http-server/handlers/pullRequestService.go`, и все три реализует один
`*postgres.Storage` — `app.New` передаёт его трижды. Добавить сигнатуру в нужный интерфейс:

```go
type UserProvider interface {
    SetIsActive(requests.UsersRequest) (models.User, error)
    DeleteUser(requests.UserDeleteRequest) (models.User, error)
}
```

**Не делать:** не создавать четвёртый интерфейс и не менять сигнатуру `handlers.New` ради
одной ручки.

### 6. Handler

В `internal/http-server/handlers/<сущность>.go`, метод на `*PullRequest`. Скелет —
в порядке, который соблюдают все существующие хендлеры:

```go
func (p *PullRequest) UsersDelete(ctx *gin.Context) {
    const op = "handlers.UsersDelete"

    log := p.log.With(slog.String("op", op))

    // 1) авторизация — до разбора тела, если ручка админская
    auth := ctx.GetHeader("Authorization")
    if auth == "" || auth != "Bearer "+adminSecret {
        log.Info("invalid token", slog.Any("token", auth))
        ctx.JSON(http.StatusUnauthorized, responses.Error(UNAUTHORIZED, "invalid token"))
        return
    }

    // 2) разбор тела
    var request requests.UserDeleteRequest
    if err := ctx.BindJSON(&request); err != nil {
        log.Error("failed to decode request body", sl.Err(err))
        ctx.JSON(http.StatusBadRequest, responses.Error(INVALID_REQUEST, "failed to decode request body"))
        return
    }

    // 3) вызов провайдера и разбор sentinel'ов — от частного к общему
    user, err := p.userProvider.DeleteUser(request)
    if err != nil {
        if errors.Is(err, storage.ErrUserNotFound) {
            log.Info("user not found", slog.String("user_id", request.UserId))
            ctx.JSON(http.StatusNotFound, responses.Error(NOT_FOUND, "user not found"))
            return
        }

        log.Error("failed to delete user", sl.Err(err))
        ctx.JSON(http.StatusInternalServerError, responses.Error(INTERNAL_ERROR, "failed to delete user"))
        return   // обязательно — в трёх старых хендлерах этого return нет, это баг
    }

    log.Info("user deleted", slog.String("user", user.UserId))
    ctx.JSON(http.StatusOK, responses.UserResponse{User: user})
}
```

Авторизация реализуется инлайн, первым блоком. Middleware в проекте нет и заводить его
ради одной ручки не нужно. Защищаются только административные операции над пользователями;
`team/*` и `pullRequest/*` открыты.

Соответствие sentinel → HTTP-статус, как в существующем коде:

| Ситуация | Статус | Код |
|---|---|---|
| не найдено | 404 | `NOT_FOUND`, `TEAM_NOT_FOUND` |
| конфликт состояния (уже есть, уже смержен, нет кандидата) | 409 | `PR_EXISTS`, `PR_MERGED`, `NO_CANDIDATE`, `NOT_ASSIGNED` |
| невалидное тело | 400 | `INVALID_REQUEST` |
| нет/неверный токен | 401 | `UNAUTHORIZED` |
| всё остальное | 500 | `INTERNAL_ERROR` |

### 7. Роут

В `cmd/pullRequestService/main.go`, в общий список рядом с соседями по сущности:

```go
router.POST("/users/delete", handlers.UsersDelete)
```

### 8. README

Дописать раздел с примером `curl`, телом ответа и перечнем кодов ошибок — так описаны
остальные ручки.

## Проверка результата

Обязательный минимум:

```sh
go build ./... && go vet ./...
make check          # конвенции AGENTS.md + правила 8-9 этого скилла
```

Затем прогнать ручку вживую, а не только собрать:

```sh
docker-compose up --build -d          # поднимает postgres + миграции + сервис на :8080
curl -X POST localhost:8080/users/delete \
  -H "Authorization: Bearer secret" -H "Content-Type: application/json" \
  -d '{"user_id":"u1"}'
```

Проверить по шагам, что:

1. успешный путь отдаёт ожидаемое тело и 200;
2. отсутствующая сущность → 404 с кодом из таблицы выше, а не 500;
3. повторный вызов ведёт себя осознанно (идемпотентно либо явной ошибкой);
4. без заголовка `Authorization` → 401, если ручка админская;
5. тело без обязательного поля → 400;
6. если в запросе есть флаг — значение `false` принимается, а не отвергается с 400
   (это и есть проверка правила про `*bool`);
7. соседние ручки не сломались: `GET /team/get` и `GET /users/getReview` отдают то же,
   что до изменений.

`make check` закрывает статические инварианты: `const op`, границы слоёв, типизированные
ответы и объявленные коды ошибок, `sl.Err`, работу внутри транзакции, нормализацию срезов,
`*bool` в DTO, совпадение пути роута с именем handler и парность миграций. Пункты 1-7 выше
он проверить не может — их прогонять руками.

## Чего не делать

- Не добавлять middleware, DI-контейнер или слой сервисов ради одной ручки — в проекте
  handler ходит в storage напрямую через провайдер-интерфейс.
- Не писать бизнес-логику в handler: выбор ревьюеров, проверки состояния и любые запросы
  к БД живут в storage.
- Не возвращать из handler данные, собранные из `request`, если storage уже вернул модель
  (echo запроса допустим только там, где так сделано исторически — `TeamAdd`).
- Не менять существующие пути роутов и имена полей JSON: они — внешний контракт.
- Не добавлять зависимости в `go.mod` ради ручки.
