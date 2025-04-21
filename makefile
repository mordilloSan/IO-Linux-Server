-include .env
-include secret.env

SERVER_PORT     ?= 8080
VITE_DEV_PORT   ?= 3000
NODE_VERSION    ?= 22
NVM_SETUP = export NVM_DIR="$$HOME/.nvm"; . "$$NVM_DIR/nvm.sh"
VERSION_FROM_ENV ?= 1.0.0
GO_BIN := $(shell which go)
AIR_BIN := $(shell which air)

default: help

check-env:
	@echo ""
	@echo "🔍 Checking .env setup..."
	@if [ -z "$(SERVER_PORT)" ]; then echo "❌ SERVER_PORT not set"; exit 1; fi
	@if [ -z "$(VITE_DEV_PORT)" ]; then echo "❌ VITE_DEV_PORT not set"; exit 1; fi
	@if [ -z "$(NODE_VERSION)" ]; then echo "❌ NODE_VERSION not set"; exit 1; fi
	@echo "✅ Environment looks good!"

.nvmrc:
	@echo $(NODE_VERSION) > .nvmrc

ensure-node: .nvmrc
	@echo ""
	@echo "📦 Ensuring Node.js $(NODE_VERSION) is available..."
	@if [ ! -d "$$HOME/.nvm" ]; then \
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash; \
	fi
	@bash -c '\
	$(NVM_SETUP); \
		nvm install $(NODE_VERSION) > /dev/null; \
		nvm use $(NODE_VERSION) > /dev/null; \
		echo "✔ Node version: $$(node -v)"; \
		echo "✔ NPM version: $$(npm -v)"; \
		echo "✔ NPX version: $$(npx -v)"; \
	'
	@echo "✅ Node.js environment ready!"

setup: .setup-complete

.setup-complete: ensure-node
	@echo ""
	@echo "📦 Installing frontend dependencies..."
	@bash -c '\
	$(NVM_SETUP); \
		cd react && npm install --silent; \
	'
	@touch .setup-complete
	@echo "✅ Frontend dependencies installed!"

dev: setup check-env
	@echo ""
	@echo "🚀 Starting dev mode (frontend + backend)..."
	@bash -c '\
	$(NVM_SETUP); \
	cd react && VITE_API_URL=http://localhost:$(SERVER_PORT) npx vite --port $(VITE_DEV_PORT) \
	' & \
	bash -c '\
	cd go-backend && \
	echo "$(SUDO_PASSWORD)" | sudo -E -S PATH="$(shell dirname $(GO_BIN)):/usr/bin:/bin" $(AIR_BIN) \
	'

lint:
	@echo "🔍 Running ESLint..."
	@bash -c '$(NVM_SETUP); \
		cd react && \
		npx eslint src --ext .js,.jsx,.ts,.tsx --fix \
	'

tsc:
	@echo "🔍 Running TypeScript type checks..."
	@bash -c '$(NVM_SETUP); \
		cd react && \
		npx tsc \
	'

test: setup check-env
	@echo ""
	@echo "📦 Running checks..."
	@$(MAKE) --no-print-directory lint
	@$(MAKE) --no-print-directory tsc
	@echo "✅ All tests done!"

build-frontend: test check-env
	@echo ""
	@echo "📦 Building frontend..."
	@bash -c '\
	$(NVM_SETUP); \
		cd react && \
		VITE_API_URL=http://localhost:$(SERVER_PORT) npx vite build && \
		echo "✅ Frontend built successfully!" \
	'

build-backend: setup
	@echo ""
	@echo "📦 Building backend..."
	@cd go-backend && \
	go build \
	-ldflags "\
		-X 'main.version=$(VERSION_FROM_ENV)' \
		-X 'main.env=$(GO_ENV)' \
		-X 'main.buildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
	-o server && \
	echo "✅ Backend built successfully!" && \
	echo "" && \
	echo "Summary:" && \
	echo "📄 Path: go-backend/server" && \
	echo "🔖 Version: $(VERSION_FROM_ENV)" && \
	echo "⏱ Build Time: $$(date -u +%Y-%m-%dT%H:%M:%SZ)" && \
	echo "📦 Size: $$(du -h server | cut -f1)" && \
	echo "🔐 SHA256: $$(shasum -a 256 server | awk '{ print $$1 }')"


build: build-frontend build-backend

binary: build-backend
	@cd go-backend && \
	GO_ENV=production SERVER_PORT=$(SERVER_PORT) ./server

prod: check-env build-frontend 
	@cd go-backend && GO_ENV=production SERVER_PORT=$(SERVER_PORT) go run .

clean:
	@rm -f .setup-complete go-backend/server || true
	@rm -rf react/node_modules go-backend/frontend || true
	@echo "🧹 Cleaned workspace."

help:
	@echo ""
	@echo "🛠️  Available commands:"
	@echo ""
	@echo "  make setup            Install frontend deps and Node.js ($(NODE_VERSION))"
	@echo "  make test             Run frontend lint + type checks"
	@echo "  make build            Run full build (frontend + backend)"
	@echo "  make build-frontend   Build React app using Vite"
	@echo "  make build-backend    Compile Go backend with version metadata"
	@echo "  make dev              Start frontend (Vite) and backend (Go) in dev mode"
	@echo "  make prod             Build react production files and run production backend"
	@echo "  make binary           Compile Go backend and run binary"
	@echo "  make clean            Remove build artifacts"
	@echo "  make check-env        Verify .env and required variables"
	@echo ""

.PHONY: all ensure-node setup dev test build-frontend build-backend build binary prod clean help lint tsc check-env