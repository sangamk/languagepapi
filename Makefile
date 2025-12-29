.PHONY: dev build build-linux run generate clean setup

# Development with hot reload
dev:
	@templ generate --watch &
	@go run ./cmd/server

# Generate templ files
generate:
	@templ generate

# Build for current platform
build: generate
	go build -o bin/server ./cmd/server

# Build for Linux (server deployment)
build-linux: generate
	GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server

# Run production build locally
run: build
	./bin/server

# Clean build artifacts
clean:
	rm -rf bin/

# Setup for first run (create .env from example)
setup:
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env - edit it with your settings"; fi

# Import Spanish words from spanish.json
import: build
	go run scripts/import_spanish.go

# Generate bridges for cards (requires GEMINI_API_KEY)
bridges:
	@if [ -z "$$GEMINI_API_KEY" ]; then echo "Error: GEMINI_API_KEY not set"; exit 1; fi
	go run -exec "env $$(cat .env | xargs)" scripts/generate_bridges.go
