.PHONY: build

default-mac: mongodb certs generate-secret build-mac

default-win: mongodb certs generate-secret build-win

default-linux: mongodb certs generate-secret build-linux

BUILD_VERSION=1.0.0
BUILD_DIR=./cmd/bin
BINARY_NAME=gophkeeper-server

GOPHKEEPER_SERVER_ADDRESS=:50051

DATABASE_PORT=27017
GOPHKEEPER_DATABASE_DSN=mongodb://localhost:$(DATABASE_PORT)
GOPHKEEPER_DATABASE_NAME=gophkeeper
DOCKER_CONTAINER_DB_NAME=test-mongo

mongodb:
	docker pull mongo
	docker run -d -p $(DATABASE_PORT):$(DATABASE_PORT) --name $(DOCKER_CONTAINER_DB_NAME) mongo:latest

GOPHKEEPER_CERT_PATH=/Users/alena/app/tls/practicum_gophkeeper_certs/localhost+2.pem
GOPHKEEPER_CERT_KEY_PATH=/Users/alena/app/tls/practicum_gophkeeper_certs/localhost+2-key.pem

certs:
	brew install mkcert  # macOS
	mkcert -install
	mkcert -cert-file $(GOPHKEEPER_CERT_PATH) -key-file $(GOPHKEEPER_CERT_KEY_PATH) localhost 127.0.0.1 ::1

generate-secret:
	@echo "Выполните ЭТУ команду для установки переменной в текущей сессии:"
	@echo "eval \$$(make exportsecret)"

exportsecret:
	@openssl rand -base64 32 | sed 's/+/-/g; s/\//_/g; s/=//g' | xargs -I {} echo 'export GOPHKEEPER_TOKEN_KEY="{}"'
	@openssl rand -base64 32 | sed 's/+/-/g; s/\//_/g; s/=//g' | xargs -I {} echo 'export GOPHKEEPER_SECRET_KEY="{}"'

show-secret:
	@bash -c '[[ -z "$$GOPHKEEPER_TOKEN_KEY" ]] && echo "GOPHKEEPER_TOKEN_KEY не установлен" || echo "GOPHKEEPER_TOKEN_KEY: $$GOPHKEEPER_TOKEN_KEY"'
	@bash -c '[[ -z "$$GOPHKEEPER_SECRET_KEY" ]] && echo "GOPHKEEPER_SECRET_KEY не установлен" || echo "GOPHKEEPER_SECRET_KEY: $$GOPHKEEPER_SECRET_KEY"'

build-mac:
	@GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.buildVersion=v$(BUILD_VERSION) -X 'main.buildDate=$$(date +'%Y/%m/%d %H:%M:%S')'" -o $(BUILD_DIR)/$(BINARY_NAME)-$(BUILD_VERSION)-darwin-arm64 ./cmd/server/main.go

build-win:
	@GOOS=windows GOARCH=amd64 go build -ldflags "-X main.buildVersion=v$(BUILD_VERSION) -X 'main.buildDate=$$(date +'%Y/%m/%d %H:%M:%S')'" -o $(BUILD_DIR)/$(BINARY_NAME)-$(BUILD_VERSION)-windows-amd64.exe ./cmd/server/main.go

build-linux:
	@GOOS=linux GOARCH=amd64 go build -ldflags "-X main.buildVersion=v$(BUILD_VERSION) -X 'main.buildDate=$$(date +'%Y/%m/%d %H:%M:%S')'" -o $(BUILD_DIR)/$(BINARY_NAME)-$(BUILD_VERSION)-linux-amd64 ./cmd/server/main.go