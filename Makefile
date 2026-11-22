APP_NAME := lynway-wms
LOCAL_BIN := $(shell go env GOPATH)/bin
BASE_APP_NAME := lynway-wms

.PHONY: swagger
swagger:
	swag init --parseDependency -g cmd/$(BASE_APP_NAME)/main.go -o docs/swagger/

.PHONY: lint
lint:
	golangci-lint run -c .golangci.yml

.PHONY: run
run:
	GOBIN=$(LOCAL_BIN) go run ./cmd/lynway-wms

.PHONY: bin-deps
bin-deps: ## Установка зависимостей
	$(info Installing project binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install github.com/swaggo/swag/cmd/swag@latest
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.7.2

