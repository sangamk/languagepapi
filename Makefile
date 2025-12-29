.PHONY: dev build build-windows run generate clean

# Development with hot reload
dev:
	@templ generate --watch &
	@go run ./cmd/server

# Generate templ files
generate:
	@templ generate

# Build everything for deployment (Linux server)
build: generate
	@mkdir -p bin/deploy
	GOOS=linux GOARCH=amd64 go build -o bin/deploy/server ./cmd/server
	@cp -r static bin/deploy/
	@cp .env.example bin/deploy/.env.example
	@if [ -f .env ]; then cp .env bin/deploy/.env; fi
	@if [ -f languagepapi.db ]; then cp languagepapi.db bin/deploy/; fi
	@echo ""
	@echo "Build complete! Deploy contents of bin/deploy/ to your server:"
	@ls -la bin/deploy/
	@echo ""
	@echo "On server: ./server"

# Build everything for Windows
build-windows: generate
	@mkdir -p bin/deploy-windows
	GOOS=windows GOARCH=amd64 go build -o bin/deploy-windows/server.exe ./cmd/server
	@cp -r static bin/deploy-windows/
	@cp .env.example bin/deploy-windows/.env.example
	@if [ -f .env ]; then cp .env bin/deploy-windows/.env; fi
	@if [ -f languagepapi.db ]; then cp languagepapi.db bin/deploy-windows/; fi
	@echo ""
	@echo "Windows build complete! Deploy contents of bin/deploy-windows/:"
	@ls -la bin/deploy-windows/
	@echo ""
	@echo "On Windows: server.exe"

# Run locally (builds for current platform)
run: generate
	go build -o bin/server ./cmd/server
	./bin/server

# Clean build artifacts
clean:
	rm -rf bin/

# Import Spanish words from spanish.json
import:
	go run scripts/import_spanish.go
