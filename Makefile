# ==============================================================================
# Arguments passing to Makefile commands
GO_INSTALLED := $(shell which go)
MG_INSTALLED := $(shell which mockgen 2> /dev/null)

BINARY=gotr
PREFIX=$$(echo $(BINARY) | tr 'a-z' 'A-Z')
POSTGRES="postgres://postgres:postgres@localhost:5432/traccar?sslmode=disable"

# Docker variables
IMAGENAME=gotrackery
PORT=5001
PORT_EXPOSE=5001

# ==============================================================================
# Install commands
install-tools:
	@echo Checking tools are installed...
ifndef MG_INSTALLED
	@echo Installing mockgen...
	@go install github.com/golang/mock/mockgen@latest
endif

# ==============================================================================
# Modules support
tidy:
	@echo Running go mod tidy...
	@go mod tidy

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
	@echo go vet
	@go vet ./...
	@echo golangci-lint
	@golangci-lint run

tests:
	@echo Running tests...
	@go test -v -race -vet=off $$(go list ./... | grep -v /temp/)

cover:
	@echo Running coverage tests...
	@go test -vet=off -coverprofile ./temp/cover.out $$(go list ./... | grep -v /temp/)
	@go tool cover -html=./temp/cover.out

# ==============================================================================
# Run commands
run: build
	@echo Running...
	@export $(PREFIX)_ADDRESS=:$(PORT)
	@./$(BINARY) tcp -p egts

# ==============================================================================
# Database commands
# make db-migrate SQL_NAME="name_of_sql_file"
db-migrate:
	@echo Creating migration...
	@migrate create -ext sql -dir ./internal/sampledb/migrations -seq -digits 8 $(SQL_NAME)

db-up:
	@echo Running UP migrations...
	@migrate -source file:./internal/sampledb/migrations -database $(POSTGRES) up

db-down:
	@echo Running DOWN migrations...
	@migrate -source file:./internal/sampledb/migrations -database $(POSTGRES) down

db-drop:
	@echo Running DROP database...
	@migrate -source file:./internal/sampledb/migrations -database $(POSTGRES) drop

# ==============================================================================
# Docker commands
docker:
	@docker build \
	-t $(IMAGENAME) .

stop:
	@docker stop $(IMAGENAME)

start:docker
	@docker run --rm --name $(IMAGENAME) \
	-p $(PORT):$(PORT_EXPOSE)/tcp -p $(PORT):$(PORT_EXPOSE)/udp \
	$(IMAGENAME) tcp -p wialonips -a :$(PORT)