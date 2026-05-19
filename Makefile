.PHONY: build test clean tidy vet

build: ## Build the static site into dist/
	go run ./cmd/gen -out dist

test: ## Run go tests
	go test -race ./...

clean: ## Remove build artifacts
	rm -rf dist

tidy: ## Tidy go.mod
	go mod tidy

vet: ## go vet
	go vet ./...
