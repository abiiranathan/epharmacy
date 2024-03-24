# Read DATABASE_URL from environment variables
# migrate -path PATH_TO_YOUR_MIGRATIONS -database YOUR_DATABASE_URL force VERSION

DATABASE_URL := $(shell echo $$DATABASE_URL)
MIGRATIONS = db/migrations
TARGET = bin/epharma

install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

create:
	@echo "Enter migration name: "; \
	read name; \
	if [ -z "$$name" ]; then \
		echo "Migration name is required"; \
		exit 1; \
	fi; \
	migrate create -ext sql -dir $(MIGRATIONS) -seq "$$name"

up:
	migrate -source file://$(MIGRATIONS) -database $(DATABASE_URL) up
force:
	migrate -source file://$(MIGRATIONS) -database $(DATABASE_URL) force $(VERSION)

down:
	migrate -source file://$(MIGRATIONS) -database $(DATABASE_URL) down 1

generate: 
	sqlc generate

build: generate
	go build -o $(TARGET) -ldflags "-s -w"
	
dev:
	air -c .air.toml

watch:
	npx tailwindcss -i ./src/input.css -o ./static/style.css --watch

format:
	./node_modules/prettier/bin/prettier.cjs --write "views/**/*.html"

.PHONY: up down generate build run create install watch