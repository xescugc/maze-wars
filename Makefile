.PHONY: help
help: ## Show this help
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/:.*##/:##/' | column -t -s '##'

.PHONY: build
build:
	@go build -v ./...

.PHONY: test
test: wasm ## Run the tests
	@xvfb-run go test ./...

.PHONY: dc-serve
dc-serve: ## Starts the server using docker-compose
	@docker-compose -f docker/docker-compose.yml -f docker/develop.yml up --build --no-deps maze-wars

.PHONY: serve
serve: wasm ## Starts the server
	@go run ./cmd/server

.PHONY: wa-build
wa-build: ## Build the wasm Game
	@env GOOS=js GOARCH=wasm go build -o ./server/assets/wasm/maze-wars.wasm ./client/wasm

.PHONY: wa-copy
wa-copy: ## Copy the 'wasm_exec.js' to execute WebAssembly binary
	@cp $$(go env GOROOT)/misc/wasm/wasm_exec.js ./server/assets/js/

.PHONY: wasm
wasm: wa-copy wa-build ## Runs all the WASM related commands to have the code ready to run
