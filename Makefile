# XFFL Development Makefile

.PHONY: help generate-events clean test build

# Default target
help: ## Show this help message
	@echo "XFFL Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

generate-events: ## Generate Go structs from AsyncAPI event specifications
	@echo "🔧 Generating Go structs from AsyncAPI specs..."
	@cd infrastructure/events && go run generate-structs.go asyncapi ../../pkg/events/generated/events.go
	@echo "✅ Generated event structs in pkg/events/generated/events.go"

validate-asyncapi: ## Validate AsyncAPI specifications
	@echo "🔍 Validating AsyncAPI specifications..."
	@if command -v asyncapi >/dev/null 2>&1; then \
		for file in infrastructure/events/asyncapi/*.yaml; do \
			echo "Validating $$file..."; \
			asyncapi validate "$$file" || exit 1; \
		done; \
		echo "✅ All AsyncAPI specs are valid"; \
	else \
		echo "⚠️  AsyncAPI CLI not found. Install with: npm install -g @asyncapi/cli"; \
		echo "Skipping validation..."; \
	fi

generate-docs: ## Generate AsyncAPI documentation
	@echo "📚 Generating AsyncAPI documentation..."
	@if command -v asyncapi >/dev/null 2>&1; then \
		mkdir -p docs/events; \
		asyncapi generate docs infrastructure/events/asyncapi/xffl-events.yaml --output docs/events/ --force-write; \
		echo "✅ Documentation generated in docs/events/"; \
	else \
		echo "❌ AsyncAPI CLI not found. Install with: npm install -g @asyncapi/cli"; \
		exit 1; \
	fi

test: ## Run all tests
	@echo "🧪 Running tests..."
	@cd pkg && go test ./...
	@cd services/afl && go test ./...
	@cd services/ffl && go test ./...
	@cd services/search && go test ./...
	@cd gateway && go test ./...
	@echo "✅ All tests passed"

build: ## Build all services
	@echo "🏗️  Building services..."
	@cd services/afl && go build -o ../../bin/afl-service cmd/server/main.go
	@cd services/ffl && go build -o ../../bin/ffl-service cmd/server/main.go
	@cd services/search && go build -o ../../bin/search-service cmd/server/main.go
	@cd gateway && go build -o ../bin/gateway main.go
	@echo "✅ All services built in bin/"

tidy: ## Clean up Go modules
	@echo "🧹 Tidying Go modules..."
	@cd pkg && go mod tidy
	@cd services/afl && go mod tidy
	@cd services/ffl && go mod tidy
	@cd services/search && go mod tidy
	@cd gateway && go mod tidy
	@echo "✅ Go modules tidied"

generate-gql: ## Generate GraphQL code for services
	@echo "🔧 Generating GraphQL code..."
	@cd services/afl && go run github.com/99designs/gqlgen generate
	@cd services/ffl && go run github.com/99designs/gqlgen generate
	@echo "✅ GraphQL code generated"

setup-zinc: ## Set up Zinc search index
	@echo "🔍 Setting up Zinc search index..."
	@curl -u admin:admin -X PUT http://localhost:4080/api/index -d @infrastructure/zinc/xffl-index-config.json -H "Content-Type: application/json"
	@echo "✅ Zinc index created"

start-services: ## Start all services (requires separate terminals)
	@echo "🚀 Starting services..."
	@echo "Run these commands in separate terminals:"
	@echo "  AFL Service:    cd services/afl && go run cmd/server/main.go"
	@echo "  FFL Service:    cd services/ffl && go run cmd/server/main.go"
	@echo "  Search Service: cd services/search && go run cmd/server/main.go"
	@echo "  Gateway:        cd gateway && go run main.go"
	@echo "  Frontend:       cd frontend/web && npm run dev"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/events/
	@echo "✅ Clean complete"

dev-setup: generate-events validate-asyncapi tidy ## Complete development setup
	@echo "✅ Development environment ready!"