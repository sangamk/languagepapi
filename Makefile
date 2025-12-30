.PHONY: dev build build-windows run generate clean sync-static sync-static-windows

# Build version (timestamp)
VERSION := $(shell powershell -NoProfile -Command "Get-Date -Format 'yyyyMMddHHmmss'")
LDFLAGS := -ldflags "-X languagepapi/internal/version.BuildVersion=$(VERSION)"

# Development with hot reload
dev:
	@templ generate --watch &
	@go run ./cmd/server

# Generate templ files
generate:
	@templ generate

# Sync static files for embedding (Unix)
sync-static:
	@cp -r static/* cmd/server/static/

# Sync static files for embedding (Windows)
sync-static-windows:
	@powershell -Command "Copy-Item -Path 'static/*' -Destination 'cmd/server/static/' -Recurse -Force"

# Build everything for deployment (Linux server)
build: generate sync-static
	@mkdir -p bin/deploy
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/deploy/server ./cmd/server
	@cp .env.example bin/deploy/.env.example
	@if [ -f .env ]; then cp .env bin/deploy/.env; fi
	@if [ -f languagepapi.db ]; then cp languagepapi.db bin/deploy/; fi
	@echo ""
	@echo "Build complete! Deploy contents of bin/deploy/ to your server:"
	@ls -la bin/deploy/
	@echo ""
	@echo "On server: ./server"

# Build for Windows (binary with embedded static files)
build-windows: generate sync-static-windows
	go build $(LDFLAGS) -o bin/server.exe ./cmd/server
	@echo Build complete: bin/server.exe [version: $(VERSION)]

# Run locally (builds for current platform)
run: generate sync-static
	go build $(LDFLAGS) -o bin/server ./cmd/server
	./bin/server

# Clean build artifacts
clean:
	rm -rf bin/

# Import Spanish words from spanish.json
import:
	go run scripts/import_spanish.go
