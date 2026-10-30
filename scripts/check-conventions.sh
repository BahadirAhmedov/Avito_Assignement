#!/usr/bin/env bash
# Проверка конвенций проекта.
#
#   AGENTS.md                              -> правила 1-7
#   .claude/skills/add-endpoint/SKILL.md   -> правила 8-9
#   AGENTS.md, раздел «Известные дефекты»  -> пороги долга
#
# Запуск:  ./scripts/check-conventions.sh
#          make check
#
# Код возврата: 0 — конвенции соблюдены, 1 — есть нарушения.

set -uo pipefail
cd "$(dirname "$0")/.."

HANDLERS=internal/http-server/handlers
STORAGE=internal/storage/postgres
REQUESTS=internal/transport/http/requests
ROUTES=cmd/pullRequestService/main.go
WIRING=$HANDLERS/pullRequestService.go

fails=0

# report <правило> <найденные нарушения>
# Вызывать без пайпа: в пайпе функция уходит в subshell и счётчик fails теряется.
report() {
	if [ -n "$2" ]; then
		printf '  FAIL  %s\n' "$1"
		printf '%s\n' "$2" | sed 's/^/          /'
		fails=$((fails + 1))
	else
		printf '  ok    %s\n' "$1"
	fi
}

echo "Конвенции AGENTS.md"
echo

# --- 1. const op первой строкой, значение = <пакет>.<Функция> --------------------
# pullRequestService.go — файл проводки, в нём только конструктор New.
out=$(awk '
	FILENAME ~ /pullRequestService\.go$/ { next }
	/^func / {
		line = $0; name = ""
		if (match(line, /^func \([^)]*\) [A-Za-z0-9_]+/)) {
			s = substr(line, RSTART, RLENGTH); sub(/^func \([^)]*\) /, "", s); name = s
		} else if (match(line, /^func [A-Za-z0-9_]+/)) {
			s = substr(line, RSTART, RLENGTH); sub(/^func /, "", s); name = s
		}
		funcline = FNR
		if ((getline nxt) <= 0) { print FILENAME ":" funcline ": " name " — нет const op"; next }
		if (nxt !~ /const op =/) { print FILENAME ":" funcline ": " name " — нет const op первой строкой"; next }
		want = (FILENAME ~ /handlers/) ? "handlers." name : "storage.postgres." name
		if (nxt !~ ("\"" want "\"")) print FILENAME ":" funcline ": " name " — op должен быть \"" want "\""
	}
' $HANDLERS/*.go $STORAGE/*.go)
report '1. const op = "<пакет>.<Функция>" первой строкой функции' "$out"

# --- 2. ошибки БД не выходят за storage, storage не знает про HTTP ---------------
out=$( { grep -rn -E 'sql\.ErrNoRows|pq\.' $HANDLERS
         grep -rn -E '"net/http"|gin-gonic|responses\.' $STORAGE; } 2>/dev/null )
report "2. sql/pq не в handlers, http/gin/responses не в storage" "$out"

# --- 3. типизированные ответы, код ошибки — константа ----------------------------
out=$( { grep -rn -E 'gin\.H|map\[string\]' internal/
         grep -rn 'ctx.JSON(http.Status' $HANDLERS | grep -v 'http.StatusOK' | grep -v 'responses.Error('
         grep -rn 'responses.Error("' $HANDLERS; } 2>/dev/null )
report "3. ответы только через responses.*, код ошибки — константа" "$out"

# Коды, переданные в responses.Error, должны быть объявлены в блоке констант.
out=$(grep -rho 'responses.Error([A-Z_][A-Z0-9_]*' $HANDLERS | sed 's/responses.Error(//' | sort -u \
	| while read -r code; do
		grep -q "^[[:space:]]*$code[[:space:]]*=" "$WIRING" || echo "$code — используется, но не объявлен в $WIRING"
	done)
report "3a. каждый код ошибки объявлен константой в pullRequestService.go" "$out"

# --- 4. log.Error всегда с sl.Err ------------------------------------------------
out=$(grep -rn 'log.Error(' $HANDLERS | grep -v 'sl.Err(' || true)
report "4. каждый log.Error содержит sl.Err(err)" "$out"

out=$(grep -rn 'slog.String("error"' $HANDLERS $STORAGE || true)
report "4a. ошибка в лог кладётся только через sl.Err" "$out"

# --- 5. внутри транзакции работаем только через tx --------------------------------
out=$(awk '
	/^func /        { fnline = FNR; body = ""; intx = 0 }
	/db\.Begin\(\)/ { intx = 1 }
	/^}/ {
		if (intx) {
			if (body !~ /defer func\(\)/)     print FILENAME ":" fnline ": транзакция без defer-обработчика"
			if (body !~ /tx\.Rollback\(\)/)   print FILENAME ":" fnline ": транзакция без Rollback"
			if (body !~ /tx\.Commit\(\)/)     print FILENAME ":" fnline ": транзакция без Commit"
			if (body ~ /s\.db\.[A-Za-z]/)     print FILENAME ":" fnline ": внутри транзакции используется s.db вместо tx"
		}
		intx = 0; body = ""
		next
	}
	# строку с s.db.Begin() из тела вычёркиваем, иначе она сама себя ловит
	{ line = $0; sub(/s\.db\.Begin\(\)/, "", line); body = body "\n" line }
' $STORAGE/*.go)
report "5. tx-функции: defer+Rollback+Commit, запросы только через tx" "$out"

# --- 6. срезы нормализуются в storage ---------------------------------------------
out=$(awk '
	/^func \(s \*Storage\)/ { fnline = FNR; body = ""; returns_slice = ($0 ~ /\[\]models\./); next }
	/^}/ {
		if (returns_slice && body !~ /== nil \{/)
			print FILENAME ":" fnline ": метод возвращает срез, но не нормализует nil в пустой срез"
		returns_slice = 0; body = ""
		next
	}
	{ body = body "\n" $0 }
' $STORAGE/*.go)
report "6. nil-срез нормализуется в []models.X{} перед возвратом" "$out"

# --- 7. bool в request-DTO только указателем ---------------------------------------
out=$(grep -rn -E '^[[:space:]]+[A-Za-z0-9_]+[[:space:]]+bool[[:space:]]+`' $REQUESTS || true)
report "7. bool в requests объявлен как *bool" "$out"

echo
echo "Конвенции skill add-endpoint"
echo

# --- 8. имя handler выводится из пути роута ----------------------------------------
# "/users/getReview" -> Users + GetReview -> UsersGetReview
out=$(grep -o 'router\.[A-Z]*("[^"]*", handlers\.[A-Za-z0-9_]*)' $ROUTES | while read -r line; do
	path=$(printf '%s' "$line" | sed 's/.*("\([^"]*\)".*/\1/')
	handler=$(printf '%s' "$line" | sed 's/.*handlers\.\([A-Za-z0-9_]*\))/\1/')
	want=$(printf '%s' "$path" | awk -F/ '{
		out = ""
		for (i = 1; i <= NF; i++) {
			if ($i == "") continue
			out = out toupper(substr($i, 1, 1)) substr($i, 2)
		}
		print out
	}')
	[ "$handler" = "$want" ] || echo "$path -> handlers.$handler, ожидалось handlers.$want"
	grep -rq "func (p \*PullRequest) $handler(" $HANDLERS || echo "$path -> handlers.$handler не объявлен в $HANDLERS"
done)
report "8. имя handler совпадает с путём роута и объявлено в handlers" "$out"

# --- 9. миграции парные и пронумерованы подряд -------------------------------------
out=$( { for up in migrations/*.up.sql; do
		[ -e "$up" ] || continue
		down="${up%.up.sql}.down.sql"
		[ -f "$down" ] || echo "$up — нет парного $down"
	done
	for down in migrations/*.down.sql; do
		[ -e "$down" ] || continue
		up="${down%.down.sql}.up.sql"
		[ -f "$up" ] || echo "$down — нет парного $up"
	done
	n=0
	for up in $(ls migrations/*.up.sql 2>/dev/null | sort -t_ -k1 -n); do
		n=$((n + 1))
		num=$(basename "$up" | cut -d_ -f1)
		[ "$num" = "$n" ] || echo "$up — нумерация с пропуском: ожидался номер $n"
	done; } )
report "9. миграции идут парами N_*.up.sql / N_*.down.sql без пропусков" "$out"

# --- Известные дефекты: пороги не должны расти ---------------------------------------
echo
echo "Известные дефекты (пороги из AGENTS.md не должны расти)"
echo

# debt <описание> <текущее> <порог> [детали]
debt() {
	if [ "$2" -gt "$3" ]; then
		printf '  FAIL  %s: %s (порог %s)\n' "$1" "$2" "$3"
		[ -n "${4:-}" ] && printf '%s\n' "$4" | sed 's/^/          /'
		fails=$((fails + 1))
	else
		printf '  ok    %s: %s (порог %s)\n' "$1" "$2" "$3"
	fi
}

no_op_list=$(grep -rn 'fmt.Errorf' $STORAGE | grep -v '"%s: %w", op' || true)
no_op=$(printf '%s' "$no_op_list" | grep -c . || true)
debt "fmt.Errorf без op" "$no_op" 1 "$no_op_list"

missing_return=$(grep -rn -A1 'responses.Error(INTERNAL_ERROR' $HANDLERS \
	| grep -A1 'ctx.JSON' | grep -c '^[^ ]*-[0-9]*-[[:space:]]*}' || true)
debt "ветки INTERNAL_ERROR без return" "$missing_return" 3

echo
if [ "$fails" -gt 0 ]; then
	echo "Нарушений: $fails. Правила — в AGENTS.md и .claude/skills/add-endpoint/SKILL.md."
	exit 1
fi
echo "Все конвенции соблюдены."
