EXAMPLE_DIR := internal/render/example

.PHONY: build test clean tidy vet example-build example-vet example-fmt example-tidy

build: ## Build the static site into dist/
	go run ./cmd/gen -out dist

test: ## Run go tests
	go test -race ./...

clean: ## Remove build artifacts
	rm -rf dist

tidy: ## Tidy go.mod (website + example)
	go mod tidy
	cd $(EXAMPLE_DIR) && go mod tidy

vet: example-vet ## go vet (website + example)
	go vet ./...

example-build: ## Compile the embedded practical example (no binary kept)
	cd $(EXAMPLE_DIR) && go build -o /dev/null .

example-vet: ## go vet the embedded practical example
	cd $(EXAMPLE_DIR) && go vet ./...

example-fmt: ## gofmt check the embedded practical example
	@test -z "$$(cd $(EXAMPLE_DIR) && gofmt -l .)" || (echo "example needs gofmt" && exit 1)
