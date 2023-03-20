#DOCKER_IMAGE="posbus-client"
#DOCKER_TAG="develop"
EXAMPLE_PORT:=0

all: build

build: dist/index.js dist/ts/index.d.ts

build/posbus.d.ts dist/index.js: bin/build_js build/wasm_exec.js build/pbc.wasm
	bin/build_js

dist/ts/index.d.ts: build/posbus.d.ts
	npm run build:types
	cp ./build/posbus.d.ts ./dist/ts

bin_pbc:
	go build -trimpath -o ./bin/pbc ./cmd/standalone
	#tinygo build -o ./bin/pbc ./cmd/standalone

bin/build_js:
	go build -trimpath -o ./bin/build_js ./cmd/build_js

build/pbc.wasm:
	GOOS=js GOARCH=wasm go build -trimpath -o ./build/pbc.wasm ./cmd/worker

build/wasm_exec.js:
	mkdir -p ./build
	cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./build/

run: bin_pbc
	./bin/pbc

run-example: bin/build_js build/pbc.wasm build/wasm_exec.js
	mkdir -p dist/
	cp example/* dist/
	bin/build_js -s -p $(EXAMPLE_PORT)

pbupdate:
	GOPROXY=direct go get -u github.com/momentum-xyz/ubercontroller/pkg/posbus@feature/musgo && go mod vendor
#test:
#	go test -v -race ./...
#

test:
	go build -trimpath -o ./bin/test ./cmd/test && ./bin/test

clean:
	rm -rf bin/ build/ dist/

.PHONY: all build run test pbupdate download install-tools
