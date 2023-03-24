#DOCKER_IMAGE="posbus-client"
#DOCKER_TAG="develop"
EXAMPLE_PORT:=0

default: help

all: js go_cli ## Build everything.

js: build_ts ## Build the javascript client.

build_ts: run_build_js
	npm run build:types

run_build_js: bin_build_js wasm go_wasm_exec
	bin/build_js

bin_build_js:
	go build -trimpath -o ./bin/build_js ./cmd/build_js

wasm:  ## Build the WASM client.
	GOOS=js GOARCH=wasm go build -trimpath -o ./build/pbc.wasm ./cmd/worker

go_cli: ## Build the golang CLI client.
	go build -trimpath -o ./bin/pbc ./cmd/standalone

go_wasm_exec:
	mkdir -p ./build
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./build/

run_example: bin_build_js go_wasm_exec
	mkdir -p dist/
	cp example/* dist/
	bin/build_js -s -p $(EXAMPLE_PORT)

pbupdate:
	GOPROXY=direct go get -u github.com/momentum-xyz/ubercontroller/pkg/posbus@develop && go mod vendor

clean:  ## Clean all generated build artifacts.
	rm -rf bin/ build/ dist/

help: ## This help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'

.PHONY: default all js build_ts run_build_js bin_build_js wasm go_cli go_wasm_exec run_example pbupdate clean help
