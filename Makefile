APP_NAME := lynway-wms
LOCAL_BIN := $(shell go env GOPATH)/bin


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

