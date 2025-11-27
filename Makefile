APP_NAME=pr-review-service
CONFIG_PATH=config.yaml

.PHONY: all up down logs ps 

all: build

# Поднять сервисы через Docker Compose
up:
	docker compose up -d --build

# Остановить сервисы
down:
	docker compose down

# Логи приложения
logs:
	docker compose logs -f app

# Проверка статуса контейнеров
ps:
	docker compose ps


