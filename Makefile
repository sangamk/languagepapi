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
	go build -o bin/server.exe ./cmd/server

# Run production build
run: build
	./bin/server.exe

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f languagepapi.db
