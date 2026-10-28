# AGENTS.md

Go-сервис распределения ревьюеров на PR. Gin + database/sql + Postgres, слои:
`cmd` → `internal/app` (проводка) → `internal/http-server/handlers` → `internal/storage/postgres`.

Здесь — только инварианты, которые действуют всегда. Пошаговая процедура добавления
HTTP-ручки вынесена в skill `add-endpoint` и подгружается по необходимости.

После любой правки: `go build ./... && go vet ./...`.

---

## 1. `const op` первой строкой функции

Каждая функция в `handlers` и `storage/postgres` начинается с `const op = "<пакет>.<Функция>"`.
Пакет пишется коротко: `handlers.TeamAdd`, `storage.postgres.SaveTeam` — не полный путь импорта.

В storage `op` идёт только в обёртку ошибки, в handlers — только в логгер:

```go
return models.User{}, fmt.Errorf("%s: %w", op, err)   // storage
log := p.log.With(slog.String("op", op))              // handlers
```

**Не делать:** не возвращать ошибку из storage без `%s: %w` с `op`; не отдавать `op` клиенту.


## 2. Ошибки БД не выходят за storage

storage переводит ошибки драйвера в sentinel'ы из `internal/storage/storage.go`
(`sql.ErrNoRows` → `storage.ErrUserNotFound`, pq `23505` → `storage.ErrTeamNameExists`).
handlers разбирают результат только через `errors.Is(err, storage.Err*)`.

**Не делать:** не упоминать `sql.ErrNoRows`, `pq.Error` и коды Postgres в `internal/http-server/`;
не импортировать `net/http`, gin или `responses` в `internal/storage/`.


## 3. Ответ клиенту — только типизированный, код ошибки — константа

Успех — именованная структура из `internal/transport/http/responses`.
Ошибка — `responses.Error(CODE, msg)`, где `CODE` — константа из блока в
`internal/http-server/handlers/pullRequestService.go`. Новая ошибка = новая константа там же.

**Не делать:** не отвечать `gin.H{...}`, `map[string]any` или голой строкой; не передавать
первым аргументом `responses.Error` строковый литерал вместо константы.


## 4. `log.Info` для 4xx, `log.Error` + `sl.Err` для 5xx

Ожидаемая бизнес-ситуация (не найдено, конфликт, невалидный токен) логируется как `Info`
с контекстом. Непредвиденная ошибка — как `Error`, и ошибка кладётся в лог только через
`sl.Err(err)` из `internal/lib/logger/sl`.

**Не делать:** не писать `slog.String("error", err.Error())` руками; не логировать `Error`
на штатные 4xx.


## 5. Транзакция — для многошаговой записи, и внутри неё только `tx`

Несколько запросов на запись → `tx, err := s.db.Begin()` плюс defer с recover/Rollback/Commit
по образцу `SaveTeam`. Одиночный statement и любые чтения — напрямую через `s.db`.
Внутри tx-функции все запросы идут через `tx`.

**Не делать:** не смешивать `s.db` и `tx` в одной функции; не оборачивать в транзакцию
одиночный `UPDATE ... RETURNING`.

**Две ловушки существующей идиомы** (см. §«Известные дефекты»): defer решает судьбу транзакции
по переменной `err`, поэтому ранний `return ..., fmt.Errorf(...)` без присваивания в `err`
даёт Commit вместо Rollback. Любой новый ранний выход **после** первой записи обязан сначала
присвоить `err`:

```go
if someBusinessConflict {
    err = fmt.Errorf("%s: %w", op, storage.ErrSomething)
    return models.PullRequest{}, err
}
```


## 6. Срез в ответе никогда не `nil`

Метод storage, возвращающий срез, перед `return` заменяет `nil` на пустой срез, чтобы в JSON
был `[]`, а не `null`. Нормализует именно storage — handler ничего не правит.

```go
if members == nil {
    members = []models.User{}
}
```

**Не делать:** не нормализовать срез в handler; не возвращать из storage `nil`-срез как «пустой».


## 7. `bool` в теле запроса объявляется как `*bool`

В структурах `internal/transport/http/requests` поле-флаг верхнего уровня — всегда указатель:

```go
IsActive *bool `json:"is_active" binding:"required"`
```

С обычным `bool` валидатор gin не отличает `false` от отсутствия поля и отвергает `false`
с 400 — ручка физически не сможет выключить флаг. Указатель различает оба случая и сохраняет
проверку обязательности.

**Не делать:** не объявлять `bool` с `binding:"required"` в request-DTO.
(Во вложенных структурах — `models.User` внутри `TeamRequest.Members` — обычный `bool` допустим:
валидатор не заходит в элементы среза без тега `dive`.)

---

## Известные дефекты — не копировать как образец

Эти места нарушают правила выше. Не воспроизводить в новом коде; при правке рядом — чинить.

- **Ответ без `return` после 5xx.** `users.go:52`, `team.go:81`, `pullrequest.go:54`: после
  `ctx.JSON(... INTERNAL_ERROR ...)` выполнение продолжается и дописывает второй ответ.
  В новом коде после любого `ctx.JSON` с ошибкой обязателен `return`.
- **Commit вместо Rollback.** В `pullrequest.go` ранние возвраты с sentinel-ошибками
  (`:39`, `:176`, `:237`) не присваивают `err`, из-за чего defer коммитит транзакцию.
  Сейчас безвредно — все они срабатывают до первой записи.
- **Потерянная ошибка коммита.** Результаты tx-функций не именованные, поэтому
  `err = tx.Commit()` в defer никуда не возвращается: сбой коммита не виден вызывающему.
  В новых tx-функциях использовать именованный результат `(err error)`.
- **`fmt.Errorf` без `op`** — `team.go:39`.

## Миграции

Пара файлов `migrations/N_init.up.sql` + `migrations/N_init.down.sql` с одним и тем же `N`
(имя `init` исторично — сохранять его для единообразия). `down` строго обратен `up`.
Применение — `make run-migrator`. Существующие файлы не редактировать, только добавлять новые.
