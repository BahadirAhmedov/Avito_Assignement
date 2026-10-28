APP1 := migrator
APP2 := pullRequestService

BIN_DIR := bin

.PHONY: build build-migrator build-service run-migrator run-service clean test lint tidy

build: build-migrator build-service

build-migrator:
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP1) ./cmd/$(APP1)

build-service:
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP2) ./cmd/$(APP2)

run-migrator:
	go run ./cmd/$(APP1)

run-service:
	go run ./cmd/$(APP2)

lint:
	golangci-lint run

# Проверка конвенций из AGENTS.md и .claude/skills/add-endpoint/SKILL.md
check:
	./scripts/check-conventions.sh

clean:
	rm -rf $(BIN_DIR)

tidy:
	go mod tidy
