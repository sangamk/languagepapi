.PHONY: dev build run generate clean

# Development with hot reload
dev:
	@templ generate --watch &
	@go run ./cmd/server

# Generate templ files
generate:
	@templ generate

# Build for production
build: generate
	@CGO_ENABLED=1 go build -o bin/server ./cmd/server

# Run production build
run: build
	@./bin/server

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -f languagepapi.db
