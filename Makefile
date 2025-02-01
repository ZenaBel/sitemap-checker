# Makefile для Sitemap Checker

# Змінні
BINARY_NAME=sitemap-checker
BUILD_DIR=cmd/sitemap-checker
DOCKER_COMPOSE=docker compose
GO=go

# Команди
.PHONY: build run docker-build docker-up docker-down clean help

# Збірка проекту
build:
	@echo "Збірка проекту..."
	$(GO) build -o $(BINARY_NAME) ./$(BUILD_DIR)

# Запуск проекту локально
run: build
	@echo "Запуск проекту локально..."
	./$(BINARY_NAME)

# Збірка Docker-образу
docker-build:
	@echo "Збірка Docker-образу..."
	$(DOCKER_COMPOSE) build

# Запуск контейнерів
docker-up: docker-build
	@echo "Запуск контейнерів..."
	$(DOCKER_COMPOSE) up -d

# Логи контейнера
logs:
	@echo "Логи контейнера..."
	$(DOCKER_COMPOSE) logs -f

# Зупинка контейнерів
docker-down:
	@echo "Зупинка контейнерів..."
	$(DOCKER_COMPOSE) down

# Очищення (видалення бінарника та тимчасових файлів)
clean:
	@echo "Очищення..."
	rm -f $(BINARY_NAME)
	rm -f errors.log
	rm -f results.json
	$(DOCKER_COMPOSE) down -v --remove-orphans

# Допомога
help:
	@echo "Доступні команди:"
	@echo "  build        - Збірка проекту"
	@echo "  run          - Запуск проекту локально"
	@echo "  docker-build - Збірка Docker-образу"
	@echo "  docker-up    - Запуск контейнерів"
	@echo "  logs         - Вивід логів контейнера"
	@echo "  docker-down  - Зупинка контейнерів"
	@echo "  clean        - Очищення (видалення бінарника та тимчасових файлів)"
	@echo "  help         - Показати цю довідку"
