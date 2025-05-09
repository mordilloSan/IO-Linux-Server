-include .env
-include secret.env

GO_VERSION      ?= 1.22.2
GO_INSTALL_DIR := $(HOME)/.go
NVM_SETUP = export NVM_DIR="$$HOME/.nvm"; . "$$NVM_DIR/nvm.sh"
GO_BIN := $(shell which go)
AIR_BIN := $(shell which air)

default: help

define check_var
	@if [ -z "$($1)" ]; then echo "❌ $1 not set. Please edit the file ".env""; exit 1; fi
endef

define check_var_sudo
	@if [ -z "$($1)" ]; then echo "❌ $1 not set. Please edit the file "secret.env""; exit 1; fi
endef

check-env: 
	@echo ""
	@echo "🔍 Checking .env setup..."
	$(call check_var,SERVER_PORT)
	$(call check_var,VITE_DEV_PORT)
	$(call check_var,GO_VERSION)
	$(call check_var,NODE_VERSION)
	$(call check_var_sudo,SUDO_PASSWORD)
	@echo "✅ Environment looks good!"

ensure-node: check-env
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

ensure-go: check-env
	@echo ""
	@echo "📦 Ensuring Go is available..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "⬇ Installing Go (no sudo)..."; \
		curl -LO https://go.dev/dl/go$(GO_VERSION).linux-amd64.tar.gz; \
		rm -rf $(GO_INSTALL_DIR); \
		mkdir -p $(GO_INSTALL_DIR); \
		tar -C $(GO_INSTALL_DIR) -xzf go$(GO_VERSION).linux-amd64.tar.gz --strip-components=1; \
		rm go$(GO_VERSION).linux-amd64.tar.gz; \
		if ! grep -q 'export PATH=$(GO_INSTALL_DIR)/bin' $$HOME/.bashrc; then \
			echo 'export PATH=$(GO_INSTALL_DIR)/bin:$$PATH' >> $$HOME/.bashrc; \
		fi; \
		echo "✔ Go installed at $(GO_INSTALL_DIR)"; \
		echo "💡 Please run 'source ~/.bashrc' or restart your terminal to use Go globally."; \
	fi
	@bash -c 'export PATH=$(GO_INSTALL_DIR)/bin:$$PATH && go version'

	@echo "✅ Go is ready!"

setup: ensure-node ensure-go
	@echo ""
	@echo "📦 Installing frontend dependencies..."
	@bash -c '\
	$(NVM_SETUP); \
		cd react && npm install --silent; \
	'
	@echo "✅ Frontend dependencies installed!"

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

test: setup
	@echo ""
	@echo "📦 Running checks..."
	@$(MAKE) --no-print-directory lint
	@$(MAKE) --no-print-directory tsc
	@echo "✅ All tests done!"

build-vite-dev: test
	@echo ""
	@echo "📦 Building frontend..."
	@bash -c '\
	$(NVM_SETUP); \
		cd react && \
		VITE_API_URL=http://localhost:$(SERVER_PORT) npx vite build && \
		echo "✅ Frontend built successfully!" \
	'

build-vite-prod: test
	@echo ""
	@echo "📦 Building frontend..."
	@bash -c '\
	$(NVM_SETUP); \
		cd react && \
		VITE_API_URL=/ npx vite build && \
		echo "✅ Frontend built successfully!" \
	'

build-backend: setup build-vite-prod
	@echo ""
	@echo "📦 Building backend..."
	@cd go-backend/cmd/server && \
	go build \
	-ldflags "\
		-X 'main.version=$(VERSION)' \
		-X 'main.env=production' \
		-X 'main.buildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
	-o server && \
	echo "✅ Backend built successfully!" && \
	echo "" && \
	echo "Summary:" && \
	echo "📄 Path: go-backend/server" && \
	echo "🔖 Version: $(VERSION)" && \
	echo "⏱ Build Time: $$(date -u +%Y-%m-%dT%H:%M:%SZ)" && \
	echo "📦 Size: $$(du -h server | cut -f1)" && \
	echo "🔐 SHA256: $$(shasum -a 256 server | awk '{ print $$1 }')"

dev: setup check-env
	@echo ""
	@echo "🚀 Starting dev mode (frontend + backend)..."
	@bash -c '\
	cd go-backend && \
	echo "$(SUDO_PASSWORD)" | sudo -E -S PATH="$(shell dirname $(GO_BIN)):/usr/bin:/bin" $(AIR_BIN) \
	' &
	@sleep 1
	@bash -c '\
	$(NVM_SETUP); \
	cd react && VITE_API_URL=http://localhost:$(SERVER_PORT) npx vite --port $(VITE_DEV_PORT) \
	'

prod: check-env build-vite-prod
	@cd go-backend/cmd/server && echo "$(SUDO_PASSWORD)" | GO_ENV=production SERVER_PORT=$(SERVER_PORT) $(GO_BIN) run .

run: build-backend
	@cd go-backend/cmd/server && \
	GO_ENV=production SERVER_PORT=$(SERVER_PORT) ./server

clean:
	@rm -f go-backend/cmd/server/server || true
	@rm -f go-backend/cmd/server/theme.json || true
	@rm -f go-backend/theme.json || true
	@rm -rf react/node_modules || true
	@rm -f react/package-lock.json || true
	@find go-backend/frontend -mindepth 1 -exec rm -rf {} + 2>/dev/null || true
	@echo "🧹 Cleaned workspace."

help:
	@echo ""
	@echo "🛠️  Available commands:"
	@echo ""
	@echo "  make check-env        Verify .env and required environment variables"
	@echo "  make setup            Install Node.js, Go and frontend dependencies"
	@echo "  make test             Run Vite linter + TypeScript type checks"
	@echo "  make dev              Start frontend (Vite) and backend (Go) in dev mode"
	@echo "  make prod             Build Vite production files and Start backend (Go) in production mode"
	@echo "  make run              Build Go binary and runs full production mode"
	@echo "  make build-vite-dev   Build frontend static files (Vite) for Go in development mode"
	@echo "  make build-vite-prod  Build frontend static files (Vite) for Go in production mode"
	@echo "  make build-backend    Build Go binary and runs it"
	@echo "  make clean            Remove build artifacts"
	@echo ""

.PHONY: all ensure-node ensure-go setup test dev prod run build-vite-dev build-vite-prod build-backend clean help lint tsc check-env