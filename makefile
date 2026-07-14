.PHONY: help start stop test lint build-docker clean migrate logs


help:
	@echo "Config Server - Available commands:"
	@echo ""
	@echo "  make start          Start Docker containers (PostgreSQL)"
	@echo "  make stop           Stop all containers"
	@echo "  make migrate        Run database migrations"
	@echo "  make test           Run all tests"
	@echo "  make lint           Lint backend code"
	@echo "  make backend-run    Start backend server (localhost:8080)"
	@echo "  make frontend-run   Start frontend dev server (localhost:5173)"
	@echo "  make build-docker   Build Docker images"
	@echo "  make clean          Remove Docker containers and volumes"
	@echo "  make logs           Show Docker logs"
	@echo ""

start:
	docker-compose up -d

stop:
	docker-compose down

migrate:
	cd backend && go run cmd/server/main.go migrate

test: 
	cd backend && go test ./... -v

lint:
	cd backend && go fmt ./...
	cd backend && go vet ./...
	cd backend && golangci-lint run

backend-run: start
	cd backend && go run cmd/server/main.go

frontend-run:
	cd frontend && npm install && npm run dev

build-docker: test
	docker build -f backend/Dockerfile -t tessera-backend:latest .
	docker build -f frontend/Dockerfile -t tessera-frontend:latest .

clean:
	docker-compose down -v

logs:
	docker-compose logs -f

.DEFAULT_GOAL := help