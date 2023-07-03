EXAMPLE_PORT:=0
OUT_DIRS=build dist bin test-results

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
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./build/

run_example: bin_build_js wasm go_wasm_exec  ## Run example server
	cp example/browser/* dist/
	bin/build_js -s -p $(EXAMPLE_PORT)

pbupdate:  ## Update the controller dependency (to latest develop branch version)
	GOPROXY=direct go get -u github.com/momentum-xyz/ubercontroller/pkg/posbus@develop && go mod vendor

test: ## Run tests
	go test -v -outputdir test-results -coverpkg ./pbc/... -coverprofile coverage.out ./...

clean:  ## Clean all generated build artifacts.
	rm -rf $(OUT_DIRS)

help: ## This help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":[^:]*?## "}; {printf "\033[38;5;69m%-30s\033[38;5;38m %s\033[0m\n", $$1, $$2}'

.PHONY: default all js build_ts run_build_js bin_build_js wasm go_cli go_wasm_exec run_example pbupdate test clean help

# 'precreate' out output directories after parsing above makefile:
$(shell mkdir -p $(OUT_DIRS))
