.DEFAULT_GOAL := start

.PHONY: help start test tidy web-install web-dev web-build build-server build run dev-api dev-web up up-d down logs clean

WEB_DIR     := web
STATIC_DIR  := $(WEB_DIR)/dist
BIN_DIR     := bin
SERVER_BIN  := $(BIN_DIR)/fuzzy-server
ADDR        ?= :8080
LOG_PATH    ?= fuzzy_journal.log
COMPOSE     := docker compose
HTTP_URL    := http://localhost$(ADDR)

help:
	@echo "Нечёткие величины — цели Makefile"
	@echo ""
	@echo "  make             — главная: собрать фронт + бэк и запустить ($(HTTP_URL))"
	@echo "  make start       — то же"
	@echo ""
	@echo "  make up          — собрать и запустить в Docker ($(HTTP_URL))"
	@echo "  make up-d        — Docker в фоне"
	@echo "  make down        — остановить контейнеры"
	@echo "  make logs        — логи контейнера"
	@echo ""
	@echo "  make test        — go test ./..."
	@echo "  make build       — только сборка (без запуска)"
	@echo ""
	@echo "  make dev-api     — hot-reload API отдельно ($(HTTP_URL))"
	@echo "  make dev-web     — hot-reload фронт (http://localhost:5173)"
	@echo ""
	@echo "  make clean       — удалить bin/, dist/, логи"

# Главная команда: React (Vite build) + Go server + раздача UI и API
start: build
	@echo ""
	@echo ">>> $(HTTP_URL) <<<"
	@echo "Журнал: $(LOG_PATH)  |  Остановка: Ctrl+C"
	@echo ""
	@STATIC_DIR=./$(STATIC_DIR) ADDR=$(ADDR) LOG_PATH=$(LOG_PATH) exec $(SERVER_BIN)

run: start

test:
	go test ./...

tidy:
	go mod tidy

web-install:
	cd $(WEB_DIR) && npm install

web-dev: web-install
	cd $(WEB_DIR) && npm run dev

web-build: web-install
	cd $(WEB_DIR) && npm run build

build-server: tidy
	@mkdir -p $(BIN_DIR)
	go build -o $(SERVER_BIN) ./cmd/server

build: web-build build-server

dev-api:
	STATIC_DIR=./$(STATIC_DIR) ADDR=$(ADDR) LOG_PATH=$(LOG_PATH) go run ./cmd/server

dev-web: web-install
	cd $(WEB_DIR) && npm run dev

up:
	$(COMPOSE) up --build

up-d:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

clean:
	rm -rf $(BIN_DIR) $(STATIC_DIR)
	rm -f fuzzy_journal.log fuzzy-app fuzzy-server
