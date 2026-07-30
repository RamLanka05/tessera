.PHONY: help start stop migrate backend-test backend-lint backend-check frontend-install frontend-lint frontend-build frontend-check compose-check docker-build ci clean logs


help:
	@echo "Config Server - Available commands:"
	@echo ""
	@echo "  make start          Start Docker containers (PostgreSQL)"
	@echo "  make stop           Stop all containers"
	@echo "  make migrate        Run database migrations"
	@echo "  make backend-test   Run backend tests"
	@echo "  make backend-lint   Format and vet backend code"
	@echo "  make backend-check  Run backend lint + tests"
	@echo "  make frontend-install Install frontend dependencies"
	@echo "  make frontend-lint  Lint frontend code"
	@echo "  make frontend-build Build frontend assets"
	@echo "  make frontend-check Run frontend lint + build"
	@echo "  make backend-run    Start backend server (localhost:8080)"
	@echo "  make frontend-run   Start frontend dev server (localhost:5173)"
	@echo "  make compose-check  Validate docker compose config"
	@echo "  make docker-build   Build Docker images if Dockerfiles exist"
	@echo "  make ci             Run the full local verification flow"
	@echo "  make clean          Remove Docker containers and volumes"
	@echo "  make logs           Show Docker logs"
	@echo ""

start:
	docker-compose up -d

stop:
	docker-compose down

migrate:
	cd backend && go run cmd/server/main.go migrate

backend-test:
	cd backend && go test ./... -v

backend-lint:
	cd backend && go fmt ./...
	cd backend && go vet ./...

backend-check: backend-lint backend-test

frontend-install:
	cd frontend/tessera-client && npm install

frontend-lint:
	cd frontend/tessera-client && npm run lint

frontend-build:
	cd frontend/tessera-client && npm run build

frontend-check: frontend-install frontend-lint frontend-build

backend-run: start
	cd backend && go run cmd/server/main.go

frontend-run:
	cd frontend/tessera-client && npm install && npm run dev

compose-check:
	docker compose config

# docker-build will only build images if the Dockerfiles exist, otherwise it will skip the build and print a message
# made this way as the Dockerfiles are not always present in the repo, and we don't want to fail the build if they are missing
docker-build:
	$(if $(wildcard backend/Dockerfile),docker build -t my-backend-image ./backend,echo Skipping backend Docker build: backend/Dockerfile not found)
	$(if $(wildcard frontend/Dockerfile),docker build -t my-frontend-image ./frontend,echo Skipping frontend Docker build: frontend/Dockerfile not found)

ci: start migrate backend-check frontend-check compose-check docker-build

clean:
	docker-compose down -v

logs:
	docker-compose logs -f

.DEFAULT_GOAL := help