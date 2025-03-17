.PHONY: help
help: ## Show this help
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/:.*##/:##/' | column -t -s '##'

.PHONY: build
build:
	@go build -v ./...

.PHONY: test
test: wasm ## Run the tests
	@xvfb-run go test g/...

.PHONY: generate
generate: ## Generates code
	@go generate ./...

.PHONY: test
ctest: wasm ## Run the tests
	@xvfb-run go test ./... -coverprofile=cover.out

.PHONY: test
cover: ## Run the cover tool
	@cover -html=cover.out

.PHONY: dc-serve
dc-serve: ## Starts the server using docker-compose
	@docker-compose -f docker/docker-compose.yml -f docker/develop.yml up --build --no-deps maze-wars

.PHONY: serve
serve: wasm ## Starts the server
	@go run ./cmd/server

.PHONY: client
client: ## Runs a client
	@go run ./cmd/client

.PHONY: wa-build
wa-build: ## Build the wasm Game
	@env GOOS=js GOARCH=wasm go build -ldflags "-X github.com/xescugc/maze-wars/client.Version=$(shell cat ./dist/metadata.json | jq .version -r | sed -e 's/^/VERSION=/;') -X github.com/xescugc/maze-wars/client.Host=https://maze-wars.yawpgames.com" -o ./wasm/main.wasm ./client/wasm

.PHONY: wa-copy
wa-copy: ## Copy the 'wasm_exec.js' to execute WebAssembly binary
	@cp $$(go env GOROOT)/misc/wasm/wasm_exec.js ./wasm/

.PHONY: wasm
wasm: wa-copy wa-build ## Runs all the WASM related commands to have the code ready to run

.PHONY: local-goreleaser
local-goreleaser: ## Generates a local release without publishing it
	@./bins/goreleaser release --snapshot --clean
	@cat ./dist/metadata.json | jq .version -r | sed -e 's/^/VERSION=/;' > ./docker/.env

.PHONY: release
release: ## Makes a full release to GitHub
	@./bins/goreleaser release --clean
	@cat ./dist/metadata.json | jq .version -r | sed -e 's/^/VERSION=/;' > ./docker/.env

.PHONY: wa-zip
wa-zip: wasm ## Zips the content of the wasm/ into a dist/isoterra-wasm.zip
	@zip -r ./dist/maze-wars-wasm.zip wasm

.PHONY: profile
profile:
	@go tool pprof --http=:8081 http://localhost:6060/debug/pprof/profile?=seconds=10

.PHONY: heap
heap:
	@go tool pprof --http=:8081 http://localhost:6060/debug/pprof/heap?=seconds=10
