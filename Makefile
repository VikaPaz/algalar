POSTGRES_PASSWORD ?= password
POSTGRES_USER ?= user
POSTGRES_DB ?= algalar

.PHONY: docker_build docker_stop run_postgres swag run oapi up_migrations down_migrations

up_migrations: ## up migrations from migrations/up.sql
	docker cp migrations/up.sql build-postgres1:/ 
	docker exec -it build-postgres-1 psql -U user -d algalar -f /up.sql

down_migrations: ## up migrations from migrations/down.sql
	docker cp migrations/down.sql build-postgres-1:/ 
	docker exec -it build-postgres-1 psql -U user -d algalar -f /down.sql
	
oapi: ## generate open-api 
	oapi-codegen  -generate chi-server,strict-server,types -package rest docs/swagger.yaml \
	> ./internal/server/rest/server.go
	
docker_compose_run:
	docker compose -f build/docker-compose.yaml up -d

run: ## run server in local machine
	go run cmd/main.go

docker_build: ## build and run app+DB in docker compose
	docker-compose -f build/docker-compose.yaml up --build

run_postgres: ## run postgres in docker
	POSTGRES_PASSWORD=${POSTGRES_PASSWORD} \
	POSTGRES_USER=${POSTGRES_USER} \
	POSTGRES_DB=${POSTGRES_DB} \
	docker-compose -f build/docker-compose.yaml up -d postgres

docker_stop: ## stop docker compose with app+DB
	docker-compose -f build/docker-compose.yaml down

.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'