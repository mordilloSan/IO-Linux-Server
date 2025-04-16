-include .env

SERVER_PORT     ?= 8080
VITE_DEV_PORT   ?= 3000
NODE_VERSION    ?= 22

.nvmrc:
	@echo $(NODE_VERSION) > .nvmrc

ensure-node: .nvmrc
	@echo "📦 Ensuring Node.js $(NODE_VERSION) is available..."
	@if [ ! -d "$$HOME/.nvm" ]; then \
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash; \
	fi
	@bash -c '\
		export NVM_DIR="$$HOME/.nvm"; \
		. "$$NVM_DIR/nvm.sh"; \
		nvm install $(NODE_VERSION) > /dev/null; \
		nvm use $(NODE_VERSION) > /dev/null; \
		echo "✔ Node version: $$(node -v)"; \
		echo "✔ NPM version: $$(npm -v)"; \
		echo "✔ NPX version: $$(npx -v)"; \
	'
	@echo "✅ Node.js environment ready!"

setup: .setup-complete

.setup-complete: ensure-node
	@echo "📦 Installing frontend dependencies..."
	@bash -c '\
		export NVM_DIR="$$HOME/.nvm"; \
		. "$$NVM_DIR/nvm.sh"; \
		nvm use $(NODE_VERSION); \
		cd react && npm install --silent; \
	'
	@touch .setup-complete
	@echo "✅ Frontend dependencies installed!"

dev: setup
	@echo "🚀 Starting dev mode (frontend + backend)..."
	@bash -c '\
		export NVM_DIR="$$HOME/.nvm"; \
		. "$$NVM_DIR/nvm.sh"; \
		nvm use $(NODE_VERSION); \
		cd react && VITE_API_URL=http://localhost:$(SERVER_PORT) npx vite --port $(VITE_DEV_PORT) \
	' & \
	bash -c '\
		cd go-backend && \
		GO_ENV=development SERVER_PORT=$(SERVER_PORT) VITE_DEV_PORT=$(VITE_DEV_PORT) go run . \
	'

build-frontend: setup
	@bash -c '\
		export NVM_DIR="$$HOME/.nvm"; \
		. "$$NVM_DIR/nvm.sh"; \
		nvm use $(NODE_VERSION); \
		cd react && \
		VITE_API_URL=http://localhost:$(SERVER_PORT) npx vite build && \
		echo "✅ Frontend built successfully!" \
	'

build-backend: setup
	@cd go-backend && \
	go build \
	-ldflags "\
		-X 'main.env=production' \
		-X 'main.version=1.0.0' \
		-X 'main.buildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
	-o server && \
	echo "✅ Backend built successfully!" && \
	echo "📄 Path: go-backend/server" && \
	echo "🔖 Version: 1.0.0" && \
	echo "⏱ Build Time: $$(date -u +%Y-%m-%dT%H:%M:%SZ)" && \
	echo "📦 Size: $$(du -h server | cut -f1)" && \
	echo "🔐 SHA256: $$(shasum -a 256 server | awk '{ print $$1 }')"

build: build-frontend build-backend

prod: build
	@cd go-backend && \
	GO_ENV=production SERVER_PORT=$(SERVER_PORT) ./server

clean:
	rm -f .setup-complete go-backend/server
	rm -rf react/node_modules go-backend/frontend
	@echo "🧹 Cleaned workspace."

.PHONY: dev setup build build-frontend build-backend prod clean ensure-node
