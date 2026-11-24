APP_NAME := lynway-wms
LOCAL_BIN := $(shell go env GOPATH)/bin

# для тестов
TEST_DB_CONTAINER = wms_postgres_test
TEST_DB_PORT = 5433
TEST_DB_USER = wms_test
TEST_DB_PASS = wms_test
TEST_DB_NAME = wms_test
TEST_DATABASE_URL ?= postgres://$(TEST_DB_USER):$(TEST_DB_PASS)@localhost:$(TEST_DB_PORT)/$(TEST_DB_NAME)?sslmode=disable



.PHONY: swagger
swagger:
	swag init --parseDependency -g cmd/$(APP_NAME)/main.go -o docs/swagger/

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

.PHONY: run
run:
	GOBIN=$(LOCAL_BIN) go run ./cmd/$(APP_NAME)

.PHONY: build
build: ## Собрать статический бинарник Community Edition (без enterprise-кода)
	GOBIN=$(LOCAL_BIN) CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" -o ./.build/$(APP_NAME)-ce $(MAIN_PATH)

.PHONY: build-enterprise
build-enterprise: ## Собрать бинарник с тегом enterprise
	GOBIN=$(LOCAL_BIN) CGO_ENABLED=0 go build -tags=enterprise -mod=vendor -ldflags="-s -w" -o ./.build/$(APP_NAME)-ee $(MAIN_PATH)

.PHONY: bin-deps
bin-deps: ## Установка зависимостей
	$(info Installing project binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@latest
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2



.PHONY: test-db-up test-db-down test-integration
test-db-up:
	@docker rm -f $(TEST_DB_CONTAINER) >/dev/null 2>&1 || true
	docker run -d --name $(TEST_DB_CONTAINER) \
		-e POSTGRES_USER=$(TEST_DB_USER) \
		-e POSTGRES_PASSWORD=$(TEST_DB_PASS) \
		-e POSTGRES_DB=$(TEST_DB_NAME) \
		-p $(TEST_DB_PORT):5432 \
		postgres:16-alpine
	@echo "Waiting for PostgreSQL to be ready..."
	until docker exec $(TEST_DB_CONTAINER) pg_isready -U $(TEST_DB_USER) -d $(TEST_DB_NAME) >/dev/null 2>&1; do sleep 1; done
	#@until docker exec $(TEST_DB_CONTAINER) psql -h 127.0.0.1 -U $(TEST_DB_USER) -d $(TEST_DB_NAME) -c "SELECT 1" >/dev/null 2>&1; do sleep 10; done
	@echo "Test PostgreSQL is ready on port $(TEST_DB_PORT)"

test-db-down:
	docker rm -f $(TEST_DB_CONTAINER) >/dev/null 2>&1 || true

test-integration: test-db-up
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -v -race ./... ; status=$$?; \
	$(MAKE) test-db-down; \
	exit $$status