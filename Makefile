.PHONY: build test serve clean tidy vet

build: ## Build the static site into dist/
	go run ./cmd/gen -out dist

test: ## Run go tests
	go test -race ./...

serve: ## Build then serve dist/ on http://localhost:8080
	go run ./cmd/gen -out dist -serve -addr :8080

clean: ## Remove build artifacts
	rm -rf dist

tidy: ## Tidy go.mod
	go mod tidy

vet: ## go vet
	go vet ./...
