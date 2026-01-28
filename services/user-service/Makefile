# ===================================
# Конфигурация
# ===================================
POSTGRES_DSN=postgres://user:password@localhost:5432/users?sslmode=disable
MIGRATION_DIR=./migrations
DOCKER_COMPOSE=docker-compose
APP_CMD=go run cmd/main.go
MIGRATE_CMD=migrate

# ===================================
# Фоновые действия
# ===================================
.PHONY: all create-migration up down migrate run clean

# -----------------------------
# 1️⃣ Создать миграцию
# -----------------------------
# Usage: make create-migration NAME=create_users_table
create-migration:
ifndef NAME
	$(error NAME is not set. Usage: make create-migration NAME=your_migration_name)
endif
	$(MIGRATE_CMD) create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)

# -----------------------------
# 2️⃣ Запустить Docker контейнеры
# -----------------------------
up:
	$(DOCKER_COMPOSE) up -d

# -----------------------------
# 3️⃣ Применить миграции
# -----------------------------
migrate:
	$(MIGRATE_CMD) -path $(MIGRATION_DIR) -database "$(POSTGRES_DSN)" up

# -----------------------------
# 4️⃣ Остановить и удалить контейнеры
# -----------------------------
down:
	$(DOCKER_COMPOSE) down

# -----------------------------
# 5️⃣ Запуск Go-сервиса
# -----------------------------
run:
	$(APP_CMD)

# -----------------------------
# 6️⃣ Полная сборка и запуск
# -----------------------------
all: up migrate run
