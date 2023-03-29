# ==============================================================================
# Arguments passing to Makefile commands
GO_INSTALLED := $(shell which go)
MG_INSTALLED := $(shell which mockgen 2> /dev/null)
SS_INSTALLED := $(shell which staticcheck 2> /dev/null)

BINARY=gotr
POSTGRES="postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable"

# Docker variables
IMAGENAME=$(BINARY)
PORT=5001
PORT_EXPOSE=5001
PG_DOCKER="postgres://postgres:postgres@host.docker.internal:5432/traccar?sslmode=disable"

# ==============================================================================
# Install commands
install-tools:
	@echo Checking tools are installed...
ifndef MG_INSTALLED
	@echo Installing mockgen...
	@go install github.com/golang/mock/mockgen@latest
endif
ifndef SS_INSTALLED
	@echo Installing staticcheck...
	@go install honnef.co/go/tools/cmd/staticcheck@latest
endif

# ==============================================================================
# Modules support
tidy:
	@echo Running go mod tidy...
	@go mod tidy
# go mod vendor

# ==============================================================================
# Build commands
gen: tidy install-tools
	@echo Running go generate...
	@go generate -x $$(go list ./... | grep -v /gen_pb/ | grep -v /googleapis/ | grep -v /pkg)

build: gen
	@echo Building...
	@go build -o ./$(BINARY) -v ./

win: gen
	@echo Building for windows...
	@GOOS=windows GOARCH=386 go build -o $(BINARY).exe ./

mac: gen
	@echo Building for mac...
	@GOOS=darwin GOARCH=amd64 go build -o $(BINARY) ./

linux: gen
	@echo Building for linux...
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY) ./

# ==============================================================================
# Test commands
lint: build
	@echo Running lints...
	@go vet ./...
	@staticcheck ./...
	@golangci-lint run

tests:
	@echo Running tests...
	@go test -v -race -vet=off $$(go list ./... | grep -v /temp/)

cover:
	@echo Running coverage tests...
	@go test -vet=off -coverprofile ./temp/cover.out $$(go list ./... | grep -v /temp/)
	@go tool cover -html=./temp/cover.out

# ==============================================================================
# Database commands
# make db-migrate SQL_NAME="name_of_sql_file"
db-migrate:
	@echo Creating migration...
	@migrate create -ext sql -dir ./internal/traccar/migrations -seq -digits 8 $(SQL_NAME)

db-up:
	@echo Running UP migrations...
	@migrate -source file:./internal/traccar/migrations -database $(POSTGRES) up

db-down:
	@echo Running DOWN migrations...
	@migrate -source file:./internal/traccar/migrations -database $(POSTGRES) down

db-drop:
	@echo Running DROP database...
	@migrate -source file:./internal/traccar/migrations -database $(POSTGRES) drop

# ==============================================================================
# Docker commands
docker-build:
	@docker build \
	--build-arg APP=$(IMAGENAME) \
	-t $(IMAGENAME) .
#	--build-arg MAIN_PATH=$(MAIN_PATH) \

stop:
	@docker stop $(IMAGENAME)

start:docker-build
	@docker run --rm --name $(IMAGENAME) \
	-e NETWORK=$(NETWORK) \
	-e ADDRESS=$(SERVER) \
	-p $(PORT):$(PORT_EXPOSE)/tcp -p $(PORT):$(PORT_EXPOSE)/udp \
	$(IMAGENAME) tcp -p wialonips -a :$(PORT)

run:
	@docker run --rm --name $(IMAGENAME) \
	-e NETWORK=$(NETWORK) \
	-e ADDRESS=$(SERVER) \
	-p $(PORT):$(PORT_EXPOSE)/tcp -p $(PORT):$(PORT_EXPOSE)/udp \
	$(IMAGENAME) tcp -p egts -a :$(PORT) --traccar $(PG_DOCKER)

ngrok:
	@ngrok tcp $(PORT_EXPOSE)

demo-srv-wialon:db-up docker-build
	@echo Running server for demo wialon
	@docker run --rm --name $(IMAGENAME) \
	-e NETWORK=$(NETWORK) \
	-e ADDRESS=$(SERVER) \
	-p $(PORT):$(PORT_EXPOSE)/tcp -p $(PORT):$(PORT_EXPOSE)/udp \
	$(IMAGENAME) tcp -p wialonips -a :$(PORT) --traccar $(PG_DOCKER)

demo-srv-egts:db-up docker-build
	@echo Running server for demo egts
	@docker run --rm --name $(IMAGENAME) \
	-e NETWORK=$(NETWORK) \
	-e ADDRESS=$(SERVER) \
	-p $(PORT):$(PORT_EXPOSE)/tcp -p $(PORT):$(PORT_EXPOSE)/udp \
	$(IMAGENAME) tcp -p egts -a :$(PORT) --traccar $(PG_DOCKER)

demo-wialon:build
	@echo Running wialonips replayer for demo
	@./$(BINARY) replay \
	-a :$(PORT_EXPOSE) \
	-p wialonips \
	-i "./temp/cap/wialon/out"

demo-egts:build
	@echo Running egts replayer for demo
	@./$(BINARY) replay \
	-a :$(PORT_EXPOSE) \
	-p egts \
	-i "./temp/cap/egts/out"